package container

import (
	"github.com/mohitkumar/orchy/server/config"
	"github.com/mohitkumar/orchy/server/persistence"
	v8 "rogchap.com/v8go"
)

type DIContiner struct {
	initialized     bool
	metadataStorage persistence.MetadataStorage
	jsvM            *v8.Isolate
}

func (p *DIContiner) setInitialized() {
	p.initialized = true
}

func NewDiContainer() *DIContiner {
	return &DIContiner{
		initialized: false,
	}
}

func (d *DIContiner) Init(conf config.Config) {
	defer d.setInitialized()
	d.jsvM = v8.NewIsolate()
}

func (d *DIContiner) GetMetadataStorage() persistence.MetadataStorage {
	if !d.initialized {
		panic("persistence not initalized")
	}
	return d.metadataStorage
}

func (d *DIContiner) GetJavaScriptVM() *v8.Isolate {
	if !d.initialized {
		panic("persistence not initalized")
	}
	return d.jsvM
}
