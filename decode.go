package main

import (
	"bufio"
	"fmt"
	"io"
	"time"
)

const (
	INDENT = 1 << iota // white space before key
	KEY                // within the characters of the key
	SPLIT              // white space between key & value
	VALUE              // within the characters of the value
)

func SplitLineMap() bufio.SplitFunc {
	state := INDENT
	offset := -1
	mark := 0
	return func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		for i, b := range data[mark:] {
			switch state {
			case INDENT:
				switch b {
				case ' ', '\t', '\r', '\n', '\b', '\f':
				default:
					offset = i
					state = KEY
				}
			case KEY:
				switch b {
				case ' ', '\t', '\b', '\f':
					state = SPLIT
					if offset == -1 {
						err = fmt.Errorf("end of key found but beginning undetermined")
						return
					} else {
						mark = 0
						offset = -1
						return i, data[offset:i], nil
					}
				case '\r', '\n':
					err = fmt.Errorf("premature line end")
					return
				default:
				}
			case SPLIT:
				switch b {
				case ' ', '\t', '\b', '\f':
				default:
					state = VALUE
					if offset == -1 {
						err = fmt.Errorf("end of offset found but beginning undetermined")
					}
				case '\r', '\n':
					err = fmt.Errorf("premature line end")
					return
				}
			case VALUE:
				switch b {
				case '\n':
					state = INDENT
					return i, data[offset : i-1], nil
				default:
					state = VALUE
					if offset == -1 {
						err = fmt.Errorf("end of offset found but beginning undetermined")
						return
					} else {
						mark = 0
						offset = -1
						return i, data[offset:i], nil
					}
				}
			}
		}
		return
	}
}

func TextDecode(r io.Reader, item *Item) error {
	scanner := bufio.NewScanner(r)
	scanner.Split(SplitLineMap())
	var key, value string
	var i, keys int
	var err error
	for keys != 7 {
		if !scanner.Scan() {
			break
		}
		key = scanner.Text()
		if !scanner.Scan() {
			return fmt.Errorf("key %s found but no value", key)
		}
		switch key {
		case "Description":
			item.Description = value
			keys |= 1 << 0
		case "Done":
			switch value {
			case "true":
				item.Done = true
			case "false":
				item.Done = false
			default:
				return fmt.Errorf("expecting true or false for %s but got %s", key, value)
			}
			keys |= 1 << 1
		case "Time":
			formats := []string{
				time.ANSIC,
				time.UnixDate,
				time.RubyDate,
				time.RFC822,
				time.RFC822Z,
				time.RFC850,
				time.RFC1123,
				time.RFC1123Z,
				time.RFC3339,
				time.RFC3339Nano,
				time.Kitchen,
			}
			for i = 0; i < len(formats); i++ {
				if item.Time, err = time.Parse(formats[i], value); err == nil {
					break
				}
			}
			if i == len(formats) {
				return fmt.Errorf("expecting date for key %s but cannot parse %s", key, value)
			}
			keys |= 1 << 2
		default:
			return fmt.Errorf("expected key %s", key)
		}
	}
	return nil
}
