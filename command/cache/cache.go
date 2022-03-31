package cache

import (
	"YaeDisk/model"
	"github.com/muesli/cache2go"
	"time"
)

var tokenCache = cache2go.Cache("tokenCache")

func InsToken(uid uint64, token model.UserAuthToken) {
	tokenCache.Add(uid, time.Hour*24*7, token)
}

func TokenCache(uid uint64, inputToken model.UserAuthToken) (bool, string) {
	if tokenCache.Exists(uid) {
		item, err := tokenCache.Value(uid)
		if err != nil {
			return false, ""
		}
		saveToken := item.Data().(model.UserAuthToken)
		if saveToken.UserID == inputToken.UserID && saveToken.UserToken == inputToken.UserToken && saveToken.IP.String() == inputToken.IP.String() {
			return true, item.LifeSpan().String()
		}
		return false, ""
	}
	return false, ""
}

func DelToken(uid uint64, inputToken model.UserAuthToken) bool {
	if tokenCache.Exists(uid) {
		item, err := tokenCache.Value(uid)
		if err != nil {
			return false
		}
		saveToken := item.Data().(model.UserAuthToken)
		if saveToken.UserID == inputToken.UserID && saveToken.UserToken == inputToken.UserToken {
			tokenCache.Delete(uid)
			return true
		}
		return false
	}
	return false
}
