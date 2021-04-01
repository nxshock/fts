package fts

import "github.com/vmihailenco/msgpack/v5"

func read(dec *msgpack.Decoder) (key string, ids []int, err error) {
	err = dec.Decode(&key)
	if err != nil {
		return "", nil, err
	}
	err = dec.Decode(&ids)
	if err != nil {
		return "", nil, err
	}

	return key, ids, nil
}

func write(enc *msgpack.Encoder, key string, ids []int) error {
	err := enc.Encode(key)
	if err != nil {
		return err
	}
	err = enc.Encode(ids)
	if err != nil {
		return err
	}

	return nil
}
