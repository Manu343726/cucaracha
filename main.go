/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"github.com/Manu343726/cucaracha/cmd"
	_ "github.com/Manu343726/cucaracha/cmd/cpu"
	_ "github.com/Manu343726/cucaracha/cmd/tools"
)

func main() {
	cmd.Execute()
}
