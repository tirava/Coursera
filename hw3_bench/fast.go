package main

import (
	"bufio"
	"fmt"
	"github.com/mailru/easyjson/jlexer"
	"io"
	"os"
	"strings"
)

//const filePath string = "./data/users.txt"

type User struct {
	Name     string   `json:"name"`
	Email    string   `json:"email"`
	Browsers []string `json:"browsers"`
}

// вам надо написать более быструю оптимальную этой функции
func FastSearch(out io.Writer) {

	seenBrowsers := make(map[string]bool)
	//foundUsers := ""
	foundUsers := make([]string, 0, 256)
	fu := ""

	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	i := -1
	user := new(User)
	var line []byte
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line = scanner.Bytes()
		i++

		//err = json.Unmarshal(line, &user)
		err = user.UnmarshalJSON(line)

		if err != nil {
			panic(err)
		}

		isAndroid := false
		isMSIE := false

		for _, browser := range user.Browsers {
			android := strings.Contains(browser, "Android")
			msie := strings.Contains(browser, "MSIE")
			if android {
				isAndroid = true
			}
			if msie {
				isMSIE = true
			}
			if android || msie {
				if _, ok := seenBrowsers[browser]; !ok {
					seenBrowsers[browser] = true
				}
			}
		}

		if !(isAndroid && isMSIE) {
			continue
		}

		email := strings.Replace(user.Email, "@", " [at] ", 1)
		//foundUsers += fmt.Sprintf("[%d] %s <%s>\n", i, user.Name, email)
		fu = fmt.Sprintf("[%d] %s <%s>\n", i, user.Name, email)
		foundUsers = append(foundUsers, fu)
	}

	fu = ""
	for _, uf := range foundUsers {
		fu += uf
	}
	//_, _ = fmt.Fprintln(out, "found users:\n"+foundUsers)
	_, _ = fmt.Fprintln(out, "found users:\n"+fu)
	_, _ = fmt.Fprintln(out, "Total unique browsers", len(seenBrowsers))

}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *User) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson3486653aDecodeCourseraHw3Bench(&r, v)
	return r.Error()
}

//// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
//func (v *User) UnmarshalEasyJSON(l *jlexer.Lexer) {
//	easyjson3486653aDecodeCourseraHw3Bench(l, v)
//}

func easyjson3486653aDecodeCourseraHw3Bench(in *jlexer.Lexer, out *User) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeString()
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "name":
			out.Name = string(in.String())
		case "email":
			out.Email = string(in.String())
		case "browsers":
			if in.IsNull() {
				in.Skip()
				out.Browsers = nil
			} else {
				in.Delim('[')
				if out.Browsers == nil {
					if !in.IsDelim(']') {
						out.Browsers = make([]string, 0, 4)
					} else {
						out.Browsers = []string{}
					}
				} else {
					out.Browsers = (out.Browsers)[:0]
				}
				for !in.IsDelim(']') {
					var v1 string
					v1 = string(in.String())
					out.Browsers = append(out.Browsers, v1)
					in.WantComma()
				}
				in.Delim(']')
			}
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
