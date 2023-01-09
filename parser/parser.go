package parser

import (
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var (
	videoSettingsSample1 *regexp.Regexp
	videoSettingsSample2 *regexp.Regexp
	startVideoPart       *regexp.Regexp
	nameVideoPart        *regexp.Regexp
)

type VideoParts struct {
	Start string
	End   string
	Name  string
}

func init() {
	var err error
	videoSettingsSample1, err = regexp.Compile(`(([0-9]*[0-9]:)+[0-9][0-9][^\\][A-zА-я| ]+)`)
	if err != nil {
		log.Fatal(err)
	}

	videoSettingsSample2, err = regexp.Compile(`([A-z| |-]* ([0-9]*[0-9]:)+[0-9][0-9])`)
	if err != nil {
		log.Fatal(err)
	}

	startVideoPart, err = regexp.Compile(`([0-9]*[0-9]:)+[0-9][0-9]`)
	if err != nil {
		log.Fatal(err)
	}

	nameVideoPart, err = regexp.Compile(`[A-zА-я]+`)
	if err != nil {
		log.Fatal(err)
	}
}

func GetVideoPartsInfo(url string) ([]VideoParts, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Print(err)
		return []VideoParts{}, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Print(err)
		return []VideoParts{}, err
	}

	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Print(err)
		return []VideoParts{}, err
	}

	videoParts, err := findVideoPartsTimecodes(string(resBody))
	if err != nil {
		log.Print(err)
		return []VideoParts{}, err
	}
	return videoParts, err
}

func findVideoPartsTimecodes(str string) ([]VideoParts, error) {
	out := videoSettingsSample1.FindAllString(str, -1)
	if len(out) == 0 {
		out = videoSettingsSample2.FindAllString(str, -1)
		if len(out) == 0 {
			return []VideoParts{}, errors.New("nothing found")
		}
	}

	hashMap := make(map[string]struct{})
	for _, s := range out {
		hashMap[s] = struct{}{}
	}

	var videoPartsArr []VideoParts
	for s, _ := range hashMap {
		timeStart := startVideoPart.FindString(s)
		nameParts := nameVideoPart.FindAllString(s, -1)
		name := strings.Replace(strings.Join(nameParts, " "), "\\", "", -1)
		videoElem := VideoParts{Start: timeStart, Name: name}
		videoPartsArr = append(videoPartsArr, videoElem)
	}

	_ = sortVideoPart(videoPartsArr)
	return videoPartsArr, nil
}

// sortVideoPart sort input array
func sortVideoPart(videoParts []VideoParts) error {
	sort.Slice(videoParts, func(i, j int) bool {
		e := regexp.MustCompile("[0-9]+")

		numbers1 := e.FindAllString(videoParts[i].Start, -1)
		numbers2 := e.FindAllString(videoParts[j].Start, -1)

		if len(numbers2) > len(numbers1) {
			return true
		} else if len(numbers2) < len(numbers1) {
			return false
		}

		for r, _ := range numbers1 {
			n, err := strconv.Atoi(numbers1[r])
			if err != nil {
				log.Println(err)
			}
			n2, err := strconv.Atoi(numbers2[r])
			if err != nil {
				log.Println(err)
			}
			if n == n2 {
				continue
			}
			return n2 > n
		}

		return false
	})
	return nil
}
