package crawler

import (
	"encoding/json"
	"errors"
	"github.com/angrymuskrat/instagram-auditor/crawler/data"
	"github.com/visheratin/unilog"
	"go.uber.org/zap"
	"reflect"
)

var (
	MsgGetNickname = "don't be able to parse response"
	ParseJsonError = errors.New(MsgGetNickname)
)

func parseNickNameResponse(body []byte, id string) (nickname string, err error) {
	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		unilog.Logger().Error(MsgGetNickname, zap.String("id", id))
		return "", err
	}

	defer func() {
		if r := recover(); r != nil {
			unilog.Logger().Error(MsgGetNickname, zap.String("id", id))
			err = ParseJsonError
		}
	}()

	nickname = proceedNicknameRequest(result)
	return nickname, nil
}

func proceedNicknameRequest(result map[string]interface{}) string {
	dataRes := result["data"].(map[string]interface{})
	user := dataRes["user"].(map[string]interface{})
	reel := user["reel"].(map[string]interface{})
	reelUser := reel["user"].(map[string]interface{})
	nickname := reelUser["username"].(string)
	return nickname
}

func parseProfile(body []byte, id string, numPosts int) (profile *data.Profile, err error) {
	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	defer func() {
		if r := recover(); r != nil {
			unilog.Logger().Error(MsgGetNickname, zap.String("id", id))
			err = ParseJsonError
		}
	}()

	profile = proceedParseProfile(result, id, numPosts)
	return  profile, nil
}

func proceedParseProfile(result map[string]interface{}, id string, numPosts int) *data.Profile {
	profile := data.Profile{}
	graphql := result["graphql"].(map[string]interface{})
	user := graphql["user"].(map[string]interface{})
	profile.Id = id
	if reflect.TypeOf(user["full_name"]) == reflect.TypeOf("") {
		profile.FullName = user["full_name"].(string)
	}
	if reflect.TypeOf(user["username"]) == reflect.TypeOf("") {
		profile.Username = user["username"].(string)
	}
	if reflect.TypeOf(user["biography"]) == reflect.TypeOf("") {
		profile.Biography = user["biography"].(string)
	}
	edge := user["edge_follow"].(map[string]interface{})
	profile.Follow = int(edge["count"].(float64))
	edge = user["edge_followed_by"].(map[string]interface{})
	profile.FollowedBy = int(edge["count"].(float64))
	profile.IsBusinessAccount = user["is_business_account"].(bool)
	profile.IsJoinedRecently = user["is_joined_recently"].(bool)
	if reflect.TypeOf(user["business_category_name"]) == reflect.TypeOf("") {
		profile.BusinessCategoryName = user["business_category_name"].(string)
	}
	if reflect.TypeOf(user["category_id"]) == reflect.TypeOf("") {
		profile.CategoryId = user["category_id"].(string)
	}
	profile.IsPrivate = user["is_private"].(bool)
	profile.IsVerified = user["is_verified"].(bool)

	if reflect.TypeOf(user["profile_pic_url"]) == reflect.TypeOf("") {
		profile.ProfilePicUrl = user["profile_pic_url"].(string)
	}
	edge = user["edge_owner_to_timeline_media"].(map[string]interface{})
	profile.PostsCount = int(edge["count"].(float64))
	edges := edge["edges"].([]interface{})

	profile.Posts = []data.Post{}
	for ind, ed := range edges {
		if ind >= numPosts {
			break
		}
		post := data.Post{}
		e := ed.(map[string]interface{})
		node := e["node"].(map[string]interface{})
		post.Shortcode = node["shortcode"].(string)

		edge = node["edge_media_to_comment"].(map[string]interface{})
		post.CommentsCount = int(edge["count"].(float64))

		edge = node["edge_liked_by"].(map[string]interface{})
		post.LikesCount = int(edge["count"].(float64))

		post.Timestamp = int(node["taken_at_timestamp"].(float64))
		post.IsVideo = node["is_video"].(bool)

		tmp := node["edge_media_to_caption"].(map[string]interface{})
		tmpArray := tmp["edges"].([]interface{})
		if len(tmpArray) > 0 {
			tmp = tmpArray[0].(map[string]interface{})
			tmp = tmp["node"].(map[string]interface{})
			post.Caption = tmp["text"].(string)
		}
		tmpArray = node["thumbnail_resources"].([]interface{})
		if len(tmpArray) > 0 {
			tmp = tmpArray[0].(map[string]interface{})
			post.ImageUrl = tmp["src"].(string)
		}
		profile.Posts = append(profile.Posts, post)
	}
	return &profile
}