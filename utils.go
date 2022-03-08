package main

import "github.com/bwmarrin/discordgo"

func reverseSlice(arr []*discordgo.Message) []*discordgo.Message {
	var output []*discordgo.Message

	for i := len(arr) - 1; i >= 0; i-- {
		output = append(output, arr[i])
	}

	return output
}
