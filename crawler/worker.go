package crawler

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/angrymuskrat/instagram-auditor/crawler/data"
	"github.com/corpix/uarand"
	"github.com/visheratin/unilog"
	"go.uber.org/zap"
	"golang.org/x/net/proxy"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

var useTor = true

type entity struct {
	id        string
	isProceed bool
}

type worker struct {
	id    int
	inCh  chan entity
	agent string
	http  http.Client
	tor   http.Client
}

func (w *worker) init(port int) {
	w.http = http.Client{
		Timeout: 30 * time.Second,
	}
	tbProxyURL, err := url.Parse("socks5://127.0.0.1:" + strconv.Itoa(port))
	if err != nil {
		return
	}
	tbDialer, err := proxy.FromURL(tbProxyURL, proxy.Direct)
	if err != nil {
		return
	}
	tbTransport := &http.Transport{
		Dial:                tbDialer.Dial,
		MaxIdleConnsPerHost: 1,
	}
	w.tor = http.Client{
		Transport: tbTransport,
		Timeout:   30 * time.Second,
	}
	w.agent = uarand.GetRandom()
}

/*func (w *worker) start() {
	for e := range w.inCh {
		w.proceedProfile(e)
		time.Sleep(2500 * time.Millisecond)
	}
}*/

func toBase64(buf []byte) string {
	return base64.StdEncoding.EncodeToString(buf)
}

func (w *worker) getProfile(nickname string, id string, numPosts int) (*data.Profile, error) {
	request := "https://www.instagram.com/" + nickname + "/?__a=1"
	body, err := w.makeRequest(request, useTor)
	if err != nil {
		return nil, err
	}
	profile, err := parseProfile(body, id, numPosts)
	if err != nil {
		return nil, err
	}
	profilePic, err := w.makeRequest(profile.ProfilePicUrl, useTor)
	if err == nil {
		profile.ProfilePic = toBase64(profilePic)
		fmt.Println(toBase64(profilePic))
	}
	for i, p := range profile.Posts {
		pic, err := w.makeRequest(p.ImageUrl, useTor)
		if err == nil {
			profile.Posts[i].Image = toBase64(pic)
		}
	}
	return profile, err
}

func (w *worker) getNickname(id string) (string, error) {
	templateAfter := "https://www.instagram.com/graphql/query/?query_hash=c9100bf9110dd6361671f113dd02e7d6&variables={%22user_id%22:%22"
	templateBefore := "%22,%20%22include_reel%22:true}"
	request := templateAfter + id + templateBefore
	body, err := w.makeRequest(request, useTor)
	if err != nil {
		return "", err
	}

	nickname, err := parseNickNameResponse(body, id)
	if err != nil {
		return "", err
	}
	return nickname, nil
}

func (w *worker) makeRequest(request string, useTor bool) ([]byte, error) {
	req, err := http.NewRequest("GET", request, nil)
	if err != nil {
		unilog.Logger().Error("unable to create request", zap.String("URL", request), zap.Error(err))
		return nil, err
	}
	req.Header.Set("user-agent", w.agent)

	var resp *http.Response
	if useTor {
		resp, err = w.tor.Do(req)
	} else {
		resp, err = w.http.Do(req)
	}
	if err != nil {
		unilog.Logger().Error("unable to make request", zap.String("URL", request), zap.Error(err))
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 429 {
		msg := fmt.Sprintf("too many requests from worker %d", w.id)
		err = errors.New(msg)
		return nil, err
	}
	if resp.StatusCode == 404 || resp.StatusCode == 500 {
		msg := "entity page was not found"
		unilog.Logger().Error(msg, zap.String("URL", request))
		err = errors.New(msg)
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)

	_ = ioutil.WriteFile("test_test.json", body, 0644)
	return body, nil
}

