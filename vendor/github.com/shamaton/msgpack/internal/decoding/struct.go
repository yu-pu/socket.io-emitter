package decoding

import (
	"encoding/binary"
	"reflect"
	"sync"

	"github.com/shamaton/msgpack/def"
)

type structCacheTypeMap struct {
	m map[string]int
}

type structCacheTypeArray struct {
	m []int
}

// struct cache map
var mapSCTM = sync.Map{}
var mapSCTA = sync.Map{}

func (d *decoder) setStruct(rv reflect.Value, offset int, k reflect.Kind) (int, error) {
	/*
		if d.isDateTime(offset) {
			dt, offset, err := d.asDateTime(offset, k)
			if err != nil {
				return 0, err
			}
			rv.Set(reflect.ValueOf(dt))
			return offset, nil
		}
	*/

	for i := range extCoders {
		if extCoders[i].IsType(offset, &d.data) {
			v, offset, err := extCoders[i].AsValue(offset, k, &d.data)
			if err != nil {
				return 0, err
			}
			rv.Set(reflect.ValueOf(v))
			return offset, nil
		}
	}

	if d.asArray {
		return d.setStructFromArray(rv, offset, k)
	}
	return d.setStructFromMap(rv, offset, k)
}

func (d *decoder) setStructFromArray(rv reflect.Value, offset int, k reflect.Kind) (int, error) {
	// get length
	l, o, err := d.sliceLength(offset, k)
	if err != nil {
		return 0, err
	}

	// find or create reference
	var scta *structCacheTypeArray
	cache, findCache := mapSCTA.Load(rv.Type())
	if !findCache {
		scta = &structCacheTypeArray{}
		for i := 0; i < rv.NumField(); i++ {
			if ok, _ := d.CheckField(rv.Type().Field(i)); ok {
				scta.m = append(scta.m, i)
			}
		}
		mapSCTA.Store(rv.Type(), scta)
	} else {
		scta = cache.(*structCacheTypeArray)
	}
	// set value
	for i := 0; i < l; i++ {
		if i < len(scta.m) {
			o, err = d.decode(rv.Field(scta.m[i]), o)
			if err != nil {
				return 0, err
			}
		} else {
			o = d.jumpOffset(o)
		}
	}
	return o, nil
}

func (d *decoder) setStructFromMap(rv reflect.Value, offset int, k reflect.Kind) (int, error) {
	// get length
	l, o, err := d.mapLength(offset, k)
	if err != nil {
		return 0, err
	}

	// find or create reference
	var sctm *structCacheTypeMap
	cache, cacheFind := mapSCTM.Load(rv.Type())
	if !cacheFind {
		sctm = &structCacheTypeMap{m: map[string]int{}}
		for i := 0; i < rv.NumField(); i++ {
			if ok, name := d.CheckField(rv.Type().Field(i)); ok {
				sctm.m[name] = i
			}
		}
		mapSCTM.Store(rv.Type(), sctm)
	} else {
		sctm = cache.(*structCacheTypeMap)
	}
	// set value if string correct
	for i := 0; i < l; i++ {
		key, o2, err := d.asString(o, k)
		if err != nil {
			return 0, err
		}
		if _, ok := sctm.m[key]; ok {
			o2, err = d.decode(rv.Field(sctm.m[key]), o2)
			if err != nil {
				return 0, err
			}
		} else {
			o2 = d.jumpOffset(o2)
		}
		o = o2
	}
	return o, nil
}

func (d *decoder) jumpOffset(offset int) int {
	code, offset := d.readSize1(offset)
	switch {
	case code == def.True, code == def.False, code == def.Nil:
		// do nothing

	case d.isPositiveFixNum(code) || d.isNegativeFixNum(code):
		// do nothing
	case code == def.Uint8, code == def.Int8:
		offset += def.Byte1
	case code == def.Uint16, code == def.Int16:
		offset += def.Byte2
	case code == def.Uint32, code == def.Int32, code == def.Float32:
		offset += def.Byte4
	case code == def.Uint64, code == def.Int64, code == def.Float64:
		offset += def.Byte8

	case d.isFixString(code):
		offset += int(code - def.FixStr)
	case code == def.Str8, code == def.Bin8:
		b, o := d.readSize1(offset)
		o += int(b)
		offset = o
	case code == def.Str16, code == def.Bin16:
		bs, o := d.readSize2(offset)
		o += int(binary.BigEndian.Uint16(bs))
		offset = o
	case code == def.Str32, code == def.Bin32:
		bs, o := d.readSize4(offset)
		o += int(binary.BigEndian.Uint32(bs))
		offset = o

	case d.isFixSlice(code):
		l := int(code - def.FixArray)
		for i := 0; i < l; i++ {
			offset = d.jumpOffset(offset)
		}
	case code == def.Array16:
		bs, o := d.readSize2(offset)
		l := int(binary.BigEndian.Uint16(bs))
		for i := 0; i < l; i++ {
			o = d.jumpOffset(o)
		}
		offset = o
	case code == def.Array32:
		bs, o := d.readSize4(offset)
		l := int(binary.BigEndian.Uint32(bs))
		for i := 0; i < l; i++ {
			o = d.jumpOffset(o)
		}
		offset = o

	case d.isFixMap(code):
		l := int(code - def.FixMap)
		for i := 0; i < l*2; i++ {
			offset = d.jumpOffset(offset)
		}
	case code == def.Map16:
		bs, o := d.readSize2(offset)
		l := int(binary.BigEndian.Uint16(bs))
		for i := 0; i < l*2; i++ {
			o = d.jumpOffset(o)
		}
		offset = o
	case code == def.Map32:
		bs, o := d.readSize4(offset)
		l := int(binary.BigEndian.Uint32(bs))
		for i := 0; i < l*2; i++ {
			o = d.jumpOffset(o)
		}
		offset = o

	case code == def.Fixext1:
		offset += def.Byte1 + def.Byte1
	case code == def.Fixext2:
		offset += def.Byte1 + def.Byte2
	case code == def.Fixext4:
		offset += def.Byte1 + def.Byte4
	case code == def.Fixext8:
		offset += def.Byte1 + def.Byte8
	case code == def.Fixext16:
		offset += def.Byte1 + def.Byte16

	case code == def.Ext8:
		b, o := d.readSize1(offset)
		o += def.Byte1 + int(b)
		offset = o
	case code == def.Ext16:
		bs, o := d.readSize2(offset)
		o += def.Byte1 + int(binary.BigEndian.Uint16(bs))
		offset = o
	case code == def.Ext32:
		bs, o := d.readSize4(offset)
		o += def.Byte1 + int(binary.BigEndian.Uint32(bs))
		offset = o

	}
	return offset
}
