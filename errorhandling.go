package main

import "fmt"

func HandleErr(err error) {
	if err != nil {
		fmt.Errorf(err.Error())
	}
}
