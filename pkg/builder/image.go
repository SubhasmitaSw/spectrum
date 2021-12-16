package builder

import (
	"fmt"
	"io/ioutil"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/cache"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

// cached will cache layers from img using the fs cache

var fs cache.Cache

func init() {
	dir, err := ioutil.TempDir("", "spectrum")
	if err != nil {
		panic(err)
	}
	fs = cache.NewFilesystemCache(dir)
}

func Pull(options Options) (v1.Image, error) {
	nameOptions := makeNameOptions(options.PullInsecure)
	ref, err := name.ParseReference(options.Base, nameOptions...)
	if err != nil {
		return nil, fmt.Errorf("parsing tag %q: %v", options.Base, err)
	}

	remoteOptions := makeRemoteOptions(options)
	img, err := remote.Image(ref, remoteOptions...)
	if err != nil {
		return nil, err
	}
	return cache.Image(img, fs), nil
}

func Push(img v1.Image, options Options) error {
	nameOptions := makeNameOptions(options.PushInsecure)
	tag, err := name.NewTag(options.Target, nameOptions...)
	if err != nil {
		return fmt.Errorf("parsing tag %q: %v", options.Target, err)
	}

	remoteOptions := makeRemoteOptions(options)
	return remote.Write(tag, img, remoteOptions...)
}

func makeNameOptions(insecure bool) (nameOptions []name.Option) {
	if insecure {
		nameOptions = append(nameOptions, name.Insecure)
	}
	return
}

func makeRemoteOptions(options Options) (remoteOptions []remote.Option) {
	if configDir := options.PushConfigDir; configDir != "" {
		keyChain := NewDirKeyChain(configDir)
		remoteOptions = append(remoteOptions, remote.WithAuthFromKeychain(keyChain))
	} else {
		remoteOptions = append(remoteOptions, remote.WithAuthFromKeychain(authn.DefaultKeychain))
	}
	if jobs := options.Jobs; jobs > 0 {
		remoteOptions = append(remoteOptions, remote.WithJobs(jobs))
	}
	return
}
