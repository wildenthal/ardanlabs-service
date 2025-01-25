package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func main() {
	var b strings.Builder
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		s := scanner.Text()
		m := make(map[string]interface{})
		err := json.Unmarshal([]byte(s), &m)
		if err != nil {
			fmt.Println(s)
			continue
		}

		b.Reset()

		b.WriteString(fmt.Sprintf("%v ", m["time"]))
		b.WriteString(fmt.Sprintf("%v ", m["level"]))
		b.WriteString(fmt.Sprintf("%v - ", m["msg"]))

		for k, v := range m {
			switch k {
			case "level", "time", "msg":
				continue
			}
			b.WriteString(k)
			b.WriteString("[")
			b.WriteString(fmt.Sprintf("%v", v))
			b.WriteString("] ")
		}
		b.WriteString("\n")

		fmt.Print(b.String())
	}
}
