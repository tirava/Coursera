package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

// !!!!!!!!!!!!!!!!!!!!!!!!!! uncomment?
//const filePath string = "./data/users.txt"

type User struct {
	Name     string   `json:"name"`
	Email    string   `json:"email"`
	Browsers []string `json:"browsers"`
}

// вам надо написать более быструю оптимальную этой функции
func FastSearch(out io.Writer) {

	r := regexp.MustCompile("@")
	//rAndroid := regexp.MustCompile("Android")
	//rMSIE := regexp.MustCompile("MSIE")
	seenBrowsers := []string{}
	uniqueBrowsers := 0
	foundUsers := ""

	//users := make([]map[string]interface{}, 0)

	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	//fileContents, err := ioutil.ReadAll(file)
	//if err != nil {
	//	panic(err)
	//}

	//lines := strings.Split(string(fileContents), "\n")
	//for _, line := range lines {

	scanner := bufio.NewScanner(file)
	i := -1
	user := new(User)
	var line []byte
	for scanner.Scan() {
		line = scanner.Bytes()
		i++

		//user := new(User)
		//err = json.Unmarshal([]byte(line), &user)
		err = json.Unmarshal(line, &user)
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
				notSeenBefore := true
				for _, item := range seenBrowsers {
					if item == browser {
						notSeenBefore = false
					}
				}
				if notSeenBefore {
					// log.Printf("SLOW New browser: %s, first seen: %s", browser, user["name"])
					seenBrowsers = append(seenBrowsers, browser)
					uniqueBrowsers++
				}
			}
		}

		if !(isAndroid && isMSIE) {
			continue
		}

		email := r.ReplaceAllString(user.Email, " [at] ")
		foundUsers += fmt.Sprintf("[%d] %s <%s>\n", i, user.Name, email)
	}

	_, _ = fmt.Fprintln(out, "found users:\n"+foundUsers)
	_, _ = fmt.Fprintln(out, "Total unique browsers", len(seenBrowsers))

}
