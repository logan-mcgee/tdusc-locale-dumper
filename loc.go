package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"unicode/utf16"
	"unicode/utf8"
)

func parseLuaFile() map[int]string {
	file, err := os.Open("SCRIPTS/LOCALISATION.LUA")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	foundTable := false

	localeMap := make(map[int]string)

	for scanner.Scan() {
		line := scanner.Text()

		if line == "}" && foundTable {
			foundTable = false
			break
		}

		if foundTable {
			trimmedLine := strings.TrimSpace(line)

			splitLine := strings.Split(trimmedLine, " = ")

			labelName := splitLine[0]
			labelIndex, err := strconv.Atoi(strings.ReplaceAll(splitLine[1], ",", ""))
			if err != nil {
				panic(err)
			}

			localeMap[labelIndex] = labelName
		}

		if line == "LocalesId = {" {
			foundTable = true
		}
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}

	return localeMap
}

func DecodeUTF16(b []byte) (string, error) {

	if len(b)%2 != 0 {
		return "", fmt.Errorf("Must have even length byte slice")
	}

	u16s := make([]uint16, 1)

	ret := &bytes.Buffer{}

	b8buf := make([]byte, 4)

	lb := len(b)
	for i := 0; i < lb; i += 2 {
		u16s[0] = uint16(b[i]) + (uint16(b[i+1]) << 8)
		r := utf16.Decode(u16s)
		n := utf8.EncodeRune(b8buf, r[0])
		ret.Write(b8buf[:n])
	}

	return ret.String(), nil
}

func parsePlocFile() []string {
	file, err := os.Open("LOCALISATION/LOCALISATION_EN.PLOC")
	if err != nil {
		panic(err)
	}

	file.Seek(4, io.SeekStart)

	locStrings := []string{}

	currentString := ""

	for true {
		utf16Characters := make([]byte, 2)

		i, err := file.Read(utf16Characters)
		if i == 0 {
			break
		}

		if err != nil {
			fmt.Println("Error reading next 2 bytes: " + err.Error())
			offset, _ := file.Seek(0, io.SeekCurrent)
			fmt.Printf("Byte position: %d\n", offset)
		}

		if utf16Characters[0] == 0x00 && utf16Characters[1] == 0x00 {
			locStrings = append(locStrings, currentString)
			currentString = ""

			continue
		}

		decodedUTF16, err := DecodeUTF16(utf16Characters)
		if err != nil {
			fmt.Println("Failed to decode utf16: " + decodedUTF16)
		}

		currentString = currentString + decodedUTF16
	}

	return locStrings
}

func main() {
	luaMap := parseLuaFile()

	locFile := parsePlocFile()
	fmt.Printf("Found %d Lua locale entries & %d locFile entries:\n", len(luaMap), len(locFile))

	foundLabels := make(map[int]string)

	for index, locString := range locFile {
		labelName, exists := luaMap[index]
		if !exists {
			fmt.Printf("no lua item found for index %d\n", index)
		}

		foundLabels[index] = labelName

		fmt.Printf("%s = %s\n", labelName, locString)
	}

	for index, labelName := range luaMap {
		if _, exists := foundLabels[index]; !exists {
			fmt.Printf("no locale text found for %s (%d)\n", labelName, index)
		}
	}
}
