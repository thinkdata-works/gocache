![CI](https://github.com/thinkdata-works/gocache/actions/workflows/ci.yml/badge.svg)

## gocache

gocache is a partitioned cache module with generic support. It uses gopromise in order to allow simultaneous fetching, where many processes may be attempting to look up a single resource.

## Requirements

gopipeline requires Go >= 1.18 since it relies on generic support

## Installation

```
$ go get github.com/thinkdata-works/gocache
```

## Quickstart

```go
package main

import (
	"io/ioutil"
	"context"
	"net/http"
	"github.com/thinkdata-works/gopipeline/pkg/gocache"
)

type User struct {
	id uint64 `json:"id"`
	login string `json:"login"`
}

type UserService struct {
	userCache *gocache.Cache[string, *User]
}

func NewUserService() *UserService {
	return &UserService{
		userCache: gocache.NewCache[string, *User]
	}
}

func (u *UserService) Get(id string) (*User, error) {
	user, err := u.userCache.Get(id, func() (*User, error) {
		r, err := http.Get(fmt.Sprintf("http://users:80/%s", id))
		if err != nil {
			return err
		}

		var u User
		err := json.NewDecoder(r.Body).Decode(&u)
		if err != nil {
			return nil, err
		}
		return &u, nil
	})
	if err != nil {
		return nil, err
	}
	return user, nil
}
```

In addition, browse the test files for example usage.

## Running tests

```
$ go test -v ./pkg/gocache
```

## License

`gocache` is licensed under the Apache 2.0 License, found in the LICENSE file.
