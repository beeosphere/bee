package core

// https://medium.com/trendyol-tech/secure-types-memory-safety-with-go-d3a20aa1e727

import (
	"math/rand"
	"time"
)

type Observer interface {
	Update(value string)
}

func CreateWatcher(name string) *Watcher {
	return &Watcher{Name: name}
}

type Watcher struct {
	Name string
}

func (c *Watcher) Update(value string) {
	// println("An event occurred: ", value)
}

type Observable struct {
	Observers []Observer
	Notifies  []time.Time
}

func (o *Observable) AddObserver(os Observer) {
	o.Observers = append(o.Observers, os)
}

func (o *Observable) HasObservers() bool {
	return len(o.Observers) > 0
}

func (o *Observable) NotifyAll(value string) {
	o.Notifies = append(o.Notifies, time.Now())

	for _, ob := range o.Observers {
		ob.Update(value)
	}
}

const KEY int = 54343

type ISecureString interface {
	Apply() ISecureString
	AddWatcher(obs Observer)
	SetKey(int)
	Set(string) ISecureString
	Get() string
	GetSelf() *SecureString
	Decrypt() []rune
	RandomizeKey()
	IsEquals(ISecureString) bool
}

type SecureString struct {
	Observable
	Key           int
	RealValue     []rune
	FakeValue     string
	Initialized   bool
	HackDetecting bool
}

func NewSecureString(value string) ISecureString {
	s := &SecureString{
		Key:           KEY,
		RealValue:     []rune(value),
		FakeValue:     value,
		Initialized:   false,
		HackDetecting: false,
	}

	s.Apply()

	return s
}

func (i *SecureString) Apply() ISecureString {
	if !i.Initialized {
		i.RealValue = i.XOR(i.RealValue, i.Key)
		i.Initialized = true
	}

	return i
}

func (i *SecureString) AddWatcher(obs Observer) {
	i.AddObserver(obs)
	i.HackDetecting = true
}

func (i *SecureString) SetKey(key int) {
	i.Key = key
}

func (i *SecureString) RandomizeKey() {
	rand.Seed(time.Now().UnixNano())

	i.RealValue = i.Decrypt()
	i.Key = rand.Intn(int(^uint(0) >> 1))
	i.RealValue = i.XOR(i.RealValue, i.Key)
}

func (i *SecureString) XOR(value []rune, key int) []rune {
	res := make([]rune, len(value))

	for i, v := range value {
		res[i] = v ^ int32(key)
	}

	return res
}

func (i *SecureString) Get() string {
	return string(i.Decrypt())
}

func (i *SecureString) GetSelf() *SecureString {
	return i
}

func (i *SecureString) Set(value string) ISecureString {
	i.RealValue = i.XOR([]rune(value), i.Key)

	if i.HackDetecting {
		i.FakeValue = value
	}

	return i
}

func (i *SecureString) Decrypt() []rune {
	if !i.Initialized {
		i.Key = KEY
		i.FakeValue = ""
		i.RealValue = i.XOR(nil, 0)
		i.Initialized = false

		return nil
	}

	res := i.XOR(i.RealValue, i.Key)

	if i.HackDetecting && string(res) != i.FakeValue {
		i.NotifyAll("hack")
	}

	return res
}

func (i *SecureString) IsEquals(o ISecureString) bool {
	if i.Key != o.GetSelf().Key {
		return string(i.XOR(i.RealValue, i.Key)) == string(i.XOR(o.GetSelf().RealValue, o.GetSelf().Key))
	}

	return string(i.RealValue) == string(o.GetSelf().RealValue)
}
