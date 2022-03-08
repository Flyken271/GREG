package main

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/marcusolsson/tui-go"
	"github.com/marcusolsson/tui-go/wordwrap"
)

type Guild struct {
	Name        string
	ID          string
	icon        string
	owner       bool
	permissions int64
	channels    []*discordgo.Channel
}

var selectedGuild Guild
var selectedChannel string

func main() {

	err := godotenv.Load()
	HandleErr(err)

	discord, err := discordgo.New(grabToken())
	HandleErr(err)

	err = discord.Open()
	HandleErr(err)

	discord.Identify.Intents = discordgo.IntentsGuildMessages

	guilds, err := discord.UserGuilds(100, "", "")
	HandleErr(err)

	guildList := tui.NewList()
	channelList := tui.NewList()
	chatArea := tui.NewVBox()

	var userGuilds []Guild

	for _, g := range guilds {
		tmpChan, err := discord.GuildChannels(g.ID)
		HandleErr(err)
		userGuilds = append(userGuilds, Guild{
			channels:    SortChannels(tmpChan),
			icon:        g.Icon,
			owner:       g.Owner,
			permissions: g.Permissions,
			ID:          g.ID,
			Name:        g.Name,
		})
	}

	if len(userGuilds) != 0 {
		for _, g := range userGuilds {
			guildList.AddItems(g.Name)
		}
	}

	guildList.OnItemActivated(func(l *tui.List) {
		channelList.RemoveItems()
		for idx, guild := range userGuilds {
			if idx == l.Selected() {
				selectedGuild = guild
				for _, channel := range guild.channels {
					if channel.Type == discordgo.ChannelTypeGuildCategory {
						channelList.AddItems("[" + channel.Name + "]")
					} else {
						channelList.AddItems(channel.Name)
					}

				}
			}
		}
	})

	channelList.OnItemActivated(func(l *tui.List) {
		for idx, val := range selectedGuild.channels {
			chatArea.Remove(idx)
			if l.Selected() == idx {
				selectedChannel = val.ID
				rawMsg, err := discord.ChannelMessages(val.ID, 100, "", "", "")
				HandleErr(err)

				for _, msg := range reverseSlice(rawMsg) {
					chatArea.Append(tui.NewHBox(
						//tui.NewLabel(msg.Timestamp.String()),
						tui.NewPadder(1, 0, tui.NewLabel(fmt.Sprintf("[%s]:", msg.Author.Username))),
						tui.NewLabel(wordwrap.WrapString(msg.Content, 60)),
						tui.NewSpacer(),
					))
				}
			}
		}
	})

	discord.AddHandler(func(s *discordgo.Session, msg *discordgo.MessageCreate) {
		if msg.ChannelID == selectedChannel {
			HandleErr(err)
			chatArea.Append(tui.NewHBox(
				tui.NewPadder(1, 0, tui.NewLabel(fmt.Sprintf("[%s]:", msg.Author.Username))),
				tui.NewLabel(wordwrap.WrapString(msg.Content, 60)),
				tui.NewSpacer(),
			))
		}
	})

	chatScroll := tui.NewScrollArea(chatArea)

	chatArea.SetBorder(true)

	chatScroll.SetAutoscrollToBottom(true)

	input := tui.NewEntry()
	input.SetSizePolicy(tui.Expanding, tui.Maximum)

	inputBox := tui.NewHBox(input)
	inputBox.SetSizePolicy(tui.Expanding, tui.Maximum)
	inputBox.SetBorder(true)

	input.OnSubmit(func(e *tui.Entry) {
		discord.ChannelMessageSend(selectedChannel, e.Text())
		input.SetText("")
	})

	chatBox := tui.NewVBox(chatScroll, inputBox)

	guildList.SetSizePolicy(tui.Maximum, tui.Preferred)
	channelList.SetSizePolicy(tui.Maximum, tui.Preferred)
	chatBox.SetSizePolicy(tui.Minimum, tui.Preferred)
	chatScroll.SetSizePolicy(tui.Maximum, tui.Preferred)

	tui.DefaultFocusChain.Set(guildList, channelList, input)

	root := tui.NewHBox(guildList, channelList, chatBox)

	ui, err := tui.New(root)
	if err != nil {
		panic(err)
	}

	ticker := time.NewTicker(5 * time.Second)
	quit := make(chan struct{})

	go func() {
		for {
			select {
			case <-ticker.C:
				ui.Repaint()
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

	ui.SetKeybinding("Esc", func() { ui.Quit(); discord.Close(); close(quit) })

	if err := ui.Run(); err != nil {
		panic(err)
	}
}
