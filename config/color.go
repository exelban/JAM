package config

import (
	"math/rand"
	"time"
)

type Color struct {
	Value string
	Used  bool
}

var (
	colors = []Color{
		{Value: "#D32F2F"},
		{Value: "#D50000"},
		{Value: "#8D6E63"},
		{Value: "#C2185B"},
		{Value: "#AB47BC"},
		{Value: "#5C6BC0"},
		{Value: "#3949AB"},
		{Value: "#1A237E"},
		{Value: "#8E24AA"},
		{Value: "#AA00FF"},
		{Value: "#7E57C2"},
		{Value: "#512DA8"},
		{Value: "#311B92"},
		{Value: "#651FFF"},
		{Value: "#3D5AFE"},
		{Value: "#1976D2"},
		{Value: "#0D47A1"},
		{Value: "#2962FF"},
		{Value: "#01579B"},
		{Value: "#4A148C"},
		{Value: "#006064"},
		{Value: "#00695C"},
		{Value: "#1B5E20"},
		{Value: "#33691E"},
		{Value: "#6D4C41"},
		{Value: "#4E342E"},
		{Value: "#DD2C00"},
		{Value: "#6A1B9A"},
		{Value: "#757575"},
		{Value: "#424242"},
		{Value: "#546E7A"},
		{Value: "#37474F"},
		{Value: "#263238"},
		{Value: "#C51162"},
	}
)

func RandomColor() string {
	allUsed := true
	for _, c := range colors {
		if !c.Used {
			allUsed = false
		}
	}

	if allUsed {
		return "#268072"
	}

	rand.Seed(time.Now().UnixNano())
	i := rand.Intn(len(colors))

	c := colors[i]
	if c.Used {
		return RandomColor()
	}

	colors[i].Used = true
	return c.Value
}
