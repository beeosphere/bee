package nectar

import "encoding/json"

func Encode(routeId string, signals []*Signal) ([]byte, error) {
	msg := &Message{
		Id:      routeId,
		Signals: signals,
	}
	return json.Marshal(msg)
}

func Decode(data []byte) (*Message, error) {
	var msg Message
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

func DecodeSignals(data []byte) ([]*Signal, string, error) {
	msg, err := Decode(data)
	if err != nil {
		return nil, "", err
	}
	return msg.Signals, msg.Id, nil
}
