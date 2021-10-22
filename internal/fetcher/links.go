package fetcher

import (
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/hi20160616/exhtml"
	"github.com/hi20160616/gears"
	"github.com/hi20160616/ms-vietnamplus/configs"
	"github.com/pkg/errors"
)

func fetchLinks() ([]string, error) {
	rt := []string{}

	for _, rawurl := range configs.Data.MS["vietnamplus"].URL {
		links, err := getLinksRss(rawurl)
		if err != nil {
			return nil, err
		}
		rt = append(rt, links...)
	}
	return rt, nil
}

// getLinksJson get links from a url that return json data.
func getLinksJson(rawurl string) ([]string, error) {
	u, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}
	raw, _, err := exhtml.GetRawAndDoc(u, 1*time.Minute)
	if err != nil {
		return nil, err
	}
	re := regexp.MustCompile(`"url":\s"(.*?)",`)
	rs := re.FindAllStringSubmatch(string(raw), -1)
	rt := []string{}
	for _, item := range rs {
		rt = append(rt, "https://"+u.Hostname()+item[1])
	}
	return gears.StrSliceDeDupl(rt), nil
}

func getLinksRss(rawurl string) ([]string, error) {
	u, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}
	links, err := exhtml.ExtractRssGuids(u.String())
	if err != nil {
		return nil, errors.WithMessagef(err, "[%s] cannot extract links from %s",
			configs.Data.MS["vietnamplus"].Title, rawurl)
	}
	links = gears.StrSliceDeDupl(links)
	links = kickOutLinksMatchPath(links, "图表新闻")
	re := regexp.MustCompile(`https://zh.vietnamplus.vn/.+/(?P<guid>\d+).vnp`)
	links2 := []string{}
	for _, link := range links {
		rs := re.ReplaceAllString(link, "https://zh.vietnamplus.vn/Utilities/Print.aspx?contentid=${guid}")
		links2 = append(links2, rs)
	}
	return links2, nil
}

func getLinks(rawurl string) ([]string, error) {
	u, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}
	if links, err := exhtml.ExtractLinks(u.String()); err != nil {
		return nil, errors.WithMessagef(err, "[%s] cannot extract links from %s",
			configs.Data.MS["vietnamplus"].Title, rawurl)
	} else {
		links = linksFilter(links, `https://tw.vietnamplus.com/\w+/\d+/.*`)
		return gears.StrSliceDeDupl(links), nil
	}
}

// kickOutLinksMatchPath will kick out the links match the path,
func kickOutLinksMatchPath(links []string, path string) []string {
	tmp := []string{}
	// path = "/" + url.QueryEscape(path) + "/"
	// path = url.QueryEscape(path)
	for _, link := range links {
		if !strings.Contains(link, path) {
			tmp = append(tmp, link)
		}
	}
	return tmp
}

func linksFilter(links []string, regex string) []string {
	flinks := []string{}
	re := regexp.MustCompile(regex)
	s := strings.Join(links, "\n")
	flinks = re.FindAllString(s, -1)
	return flinks
}

func kickOut(links []string, regex string) []string {
	blackList := linksFilter(links, regex)
	return gears.StrSliceDiff(links, blackList)
}
