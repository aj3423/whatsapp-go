package def

import (
	"math/rand"
	"strings"

	"arand"
)

// Android 7: 9fc6ef6efc99b933c5e2d8fcf4f68955
var Ja3_Emu = `771,49195-49196-52393-49199-49200-52392-158-159-49161-49162-49171-49172-51-57-156-157-47-53,65281-0-23-35-13-16-11-10,23-24-25,0`

// Android 8/9: d8c87b9bfde38897979e41242626c2f3
var Ja3_WhiteMi6x = `771,49195-49196-52393-49199-49200-52392-49161-49162-49171-49172-156-157-47-53,65281-0-23-35-13-5-16-11-10,29-23-24,0`

// Android 10/11: 9b02ebd3a43b62d825e1ac605b621dc8
var Ja3_Android_10 = `771,4865-4866-4867-49195-49196-52393-49199-49200-52392-49161-49162-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-51-45-43-21,29-23-24,0`

const (
	Ja3Type_Default   int8 = 1 // the default, White Mi6x
	Ja3Type_RandomGen int8 = 2
	Ja3Type_Custom    int8 = 3
)

func NewRandomJa3Config() string {
	parts := strings.Split(Ja3_WhiteMi6x, ",")
	extension_str := parts[1]

	extensions := strings.Split(extension_str, "-")

	fixed := extensions[0:8] // fixed size: 8
	rest := extensions[8:]

	for _, r := range rest {
		if arand.Bool() {
			fixed = append(fixed, r)
		}
	}

	rand.Shuffle(len(fixed), func(i, j int) { fixed[i], fixed[j] = fixed[j], fixed[i] })
	parts[1] = strings.Join(fixed, "-")

	return strings.Join(parts, ",")
}
