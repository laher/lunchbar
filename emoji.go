package main

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"image/png"
	"path/filepath"
	"strings"

	ico "github.com/Kodeworks/golang-image-ico"
	"github.com/apex/log"
	"github.com/matryer/xbar/pkg/plugins"
)

// TODO include emoji for windows builds ...
// hold the emojis
//go:embed emoji/*
var emoji embed.FS

func GetEmoji(icon string) ([]byte, error) {
	return emoji.ReadFile(fmt.Sprintf("emoji/%s.png", icon))
}

func getNames() map[string]string {
	// TODO cache this
	em := ListEmoji()
	names := map[string]string{}
	for _, e := range em {
		emojise := plugins.Emojize(":" + e + ":")
		names[emojise] = e
	}
	// todo check these exist in there ...
	names["0"] = "zero"
	names["1"] = "one"
	names["2"] = "two"
	names["3"] = "three"
	names["4"] = "four"
	names["5"] = "five"
	names["6"] = "six"
	names["7"] = "seven"
	names["8"] = "eight"
	names["9"] = "nine"
	names["v"] = "v"
	names["⚓"] = "anchor"
	names["⚠"] = "warning"
	names["↑"] = "arrow_up"
	return names
}

func getEmojicon(in []byte) ([]byte, error) {
	img, err := png.Decode(bytes.NewReader(in))
	if err != nil {
		return nil, err
	}
	w := &bytes.Buffer{}
	err = ico.Encode(w, img)
	if err != nil {
		log.Errorf("cant encode ico", err)
		return nil, err
	}
	return w.Bytes(), nil
}

func GetIconForChar(ch string) ([]byte, error) {
	icon, ok := getNames()[ch]
	if !ok {
		for k := range getNames() {
			fmt.Println(k)
		}
		return nil, errors.New("emoji not found")
	}
	return GetEmoji(icon)
}

func ListEmoji() []string {
	res := []string{}
	rd, err := emoji.ReadDir("emoji")
	if err != nil {
		return res
	}
	for _, r := range rd {
		n := r.Name()
		n = filepath.Base(n)
		if strings.HasSuffix(n, ".png") {
			res = append(res, n[:len(n)-4])
		}
	}
	return res
}
