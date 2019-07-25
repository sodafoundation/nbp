// Copyright 2019 The OpenSDS Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sanity

import (
	"errors"
	"fmt"
	"reflect"
)

func structCopy(source, target interface{}) error {
	src, tgt, err := checkSrcAndTgtParam(source, target)
	if err != nil {
		return err
	}

	if src.Kind() != reflect.Struct && tgt.Kind() != reflect.Struct {
		return errors.New("source and target must be all struct")
	}

	for i := 0; i < src.NumField(); i++ {
		srcFiled := src.Type().Field(i)
		srcFiledName := srcFiled.Name
		tgtFiled := tgt.FieldByName(srcFiledName)

		if !tgtFiled.CanSet() {
			return fmt.Errorf("%s field is unaddressable", srcFiledName)
		}
		tgtFiled.Set(src.Field(i))
	}

	return nil
}

func structListCopy(source, target interface{}) error {
	src, tgt, err := checkSrcAndTgtParam(source, target)
	if err != nil {
		return err
	}

	if src.Kind() != reflect.Slice && tgt.Kind() != reflect.Slice {
		return errors.New("source and target must be all slice")
	}

	for i := 0; i < src.Len(); i++ {
		tgt.Set(reflect.Append(tgt, src.Index(i)))
	}

	return nil
}

func checkSrcAndTgtParam(source, target interface{}) (reflect.Value, reflect.Value, error) {
	src := reflect.ValueOf(source)
	tgt := reflect.ValueOf(target)

	if src.Kind() == reflect.Ptr {
		src = src.Elem()
	}

	if tgt.Kind() == reflect.Ptr {
		tgt = tgt.Elem()
	}

	if !tgt.CanSet() {
		return reflect.ValueOf(nil), reflect.ValueOf(nil), fmt.Errorf("target is unaddressable")
	}

	return src, tgt, nil
}
