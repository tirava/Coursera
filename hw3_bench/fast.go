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
	for scanner.Scan() {
		line := scanner.Text()
		i++

		//user := make(map[string]interface{}, 1)
		//fmt.Printf("%v %v\n", err, line)
		user := new(User)
		err = json.Unmarshal([]byte(line), &user)
		if err != nil {
			panic(err)
		}
		//fmt.Println(user)
		//	users = append(users, user)
		//}
		//
		//for i, user := range users {

		isAndroid := false
		isMSIE := false

		//browsers, ok := user["browsers"].([]interface{})
		//if !ok {
		//	// log.Println("cant cast browsers")
		//	continue
		//}

		//for _, browserRaw := range browsers {
		for _, browser := range user.Browsers {
			//browser, ok := browserRaw.(string)
			//if !ok {
			//	// log.Println("cant cast browser to string")
			//	continue
			//}
			if strings.Contains(browser, "Android") {
				//if rAndroid.MatchString(browser) {
				//if ok, err := regexp.MatchString("Android", browser); ok && err == nil {
				isAndroid = true
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

		//for _, browserRaw := range browsers {
		for _, browser := range user.Browsers {
			//browser, ok := browserRaw.(string)
			//if !ok {
			//	// log.Println("cant cast browser to string")
			//	continue
			//}
			if strings.Contains(browser, "MSIE") {
				//if rMSIE.MatchString(browser) {
				//if ok, err := regexp.MatchString("MSIE", browser); ok && err == nil {
				isMSIE = true
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

		// log.Println("Android and MSIE user:", user["name"], user["email"])
		//email := r.ReplaceAllString(user["email"].(string), " [at] ")
		email := r.ReplaceAllString(user.Email, " [at] ")
		//foundUsers += fmt.Sprintf("[%d] %s <%s>\n", i, user["name"], email)
		foundUsers += fmt.Sprintf("[%d] %s <%s>\n", i, user.Name, email)
	}

	_, _ = fmt.Fprintln(out, "found users:\n"+foundUsers)
	_, _ = fmt.Fprintln(out, "Total unique browsers", len(seenBrowsers))

	//fmt.Println(out)
}

//func parseUser(user map[string]interface{})  {
//
//}
