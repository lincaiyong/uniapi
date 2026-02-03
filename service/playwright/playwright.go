package playwright

import (
	"github.com/lincaiyong/erro"
	"github.com/lincaiyong/log"
	"github.com/playwright-community/playwright-go"
	"regexp"
	"strings"
)

func GetCookies(url, element, cookies string) (result []string, err error) {
	defer erro.Recover(func(e error) { err = e })

	log.InfoLog("start playwright")
	pw := erro.Check1(playwright.Run())
	defer func() { _ = pw.Stop() }()

	log.InfoLog("start browser")
	userDataDir := "/tmp/user-data"
	browser := erro.Check1(pw.Chromium.LaunchPersistentContext(userDataDir, playwright.BrowserTypeLaunchPersistentContextOptions{
		Channel:  playwright.String("chrome"),
		Headless: playwright.Bool(false),
		Args: []string{
			"--no-first-run",
			"--no-default-browser-check",
		},
	}))
	defer func() { _ = browser.Close() }()

	log.InfoLog("open page & goto url")
	page := erro.Check1(browser.NewPage())
	erro.Check1(page.Goto(url))

	log.InfoLog("wait for element")
	erro.Check1(page.WaitForSelector(element))

	log.InfoLog("read cookies: %s", cookies)
	m := make(map[string]string)
	for _, name := range strings.Split(cookies, ",") {
		m[name] = ""
	}
	cs := erro.Check1(browser.Cookies())
	for _, cookie := range cs {
		if _, ok := m[cookie.Name]; ok {
			m[cookie.Name] = cookie.Value
		}
	}
	for _, name := range strings.Split(cookies, ",") {
		result = append(result, m[name])
	}
	log.InfoLog("done")
	return
}

func GetHeader(url string, urlPattern *regexp.Regexp, header string) (result string, err error) {
	defer erro.Recover(func(e error) { err = e })
	log.InfoLog("start playwright")
	pw := erro.Check1(playwright.Run())
	defer func() { _ = pw.Stop() }()

	log.InfoLog("start browser")
	userDataDir := "/tmp/user-data"
	browser := erro.Check1(pw.Chromium.LaunchPersistentContext(userDataDir, playwright.BrowserTypeLaunchPersistentContextOptions{
		Channel:  playwright.String("chrome"),
		Headless: playwright.Bool(false),
		Args: []string{
			"--no-first-run",
			"--no-default-browser-check",
		},
	}))
	defer func() { _ = browser.Close() }()

	log.InfoLog("open page & goto url")
	page := erro.Check1(browser.NewPage())
	erro.Check1(page.Goto(url))

	log.InfoLog("wait for request")
	resp := erro.Check1(page.ExpectResponse(urlPattern, func() error { return nil }))
	
	headers := resp.Request().Headers()
	log.InfoLog("read headers: %s", headers)
	return headers[header], nil
}
