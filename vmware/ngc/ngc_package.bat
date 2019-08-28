:: Copyright 2019 The OpenSDS Authors.
::
:: Licensed under the Apache License, Version 2.0 (the "License");
:: you may not use this file except in compliance with the License.
:: You may obtain a copy of the License at
::
::     http://www.apache.org/licenses/LICENSE-2.0
::
:: Unless required by applicable law or agreed to in writing, software
:: distributed under the License is distributed on an "AS IS" BASIS,
:: WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
:: See the License for the specific language governing permissions and
:: limitations under the License.
@echo on 
echo step one....
cd NGC-AdapterManager/
call ant -buildfile build-java.xml
echo step two....
cd ../NGC-Plugin/makefile/
call ant -buildfile ngc_part0.xml
echo step three....
cd ../../NGC-Register/
mvn package