# Copyright 2016-2017 VMware, Inc. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#	http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License

*** Settings ***
Documentation  Test 1-32 - Docker plugin
Resource  ../../resources/Util.robot
Suite Setup  Install VIC Appliance To Test Server  certs=${false}
Suite Teardown  Cleanup VIC Appliance On Test Server

*** Test Cases ***
Docker plugin install
    ${rc}  ${output}=  Run And Return Rc And Output  docker1.13 %{VCH-PARAMS} plugin install vieux/sshfs
    Should Be Equal As Integers  ${rc}  1
    Should Contain  ${output}  does not yet support plugins

Docker plugin create
    Run  echo '{}' > config.json
    ${rc}  ${output}=  Run And Return Rc And Output  docker1.13 %{VCH-PARAMS} plugin create test-plugin .
    Should Be Equal As Integers  ${rc}  1
    Should Contain  ${output}  does not yet support plugins

Docker plugin enable
    ${rc}  ${output}=  Run And Return Rc And Output  docker1.13 %{VCH-PARAMS} plugin enable test-plugin
    Should Be Equal As Integers  ${rc}  1
    Should Contain  ${output}  does not yet support plugins

Docker plugin disable
    ${rc}  ${output}=  Run And Return Rc And Output  docker1.13 %{VCH-PARAMS} plugin disable test-plugin
    Should Be Equal As Integers  ${rc}  1
    Should Contain  ${output}  does not yet support plugins

Docker plugin inspect
    ${rc}  ${output}=  Run And Return Rc And Output  docker1.13 %{VCH-PARAMS} plugin inspect test-plugin
    Should Be Equal As Integers  ${rc}  1
    Should Contain  ${output}  does not yet support plugins

Docker plugin ls
    ${rc}  ${output}=  Run And Return Rc And Output  docker1.13 %{VCH-PARAMS} plugin ls
    Should Be Equal As Integers  ${rc}  1
    Should Contain  ${output}  does not yet support plugins

Docker plugin push
    ${rc}  ${output}=  Run And Return Rc And Output  docker1.13 %{VCH-PARAMS} plugin push test-plugin
    Should Be Equal As Integers  ${rc}  1
    Should Contain  ${output}  does not yet support plugins

Docker plugin rm
    ${rc}  ${output}=  Run And Return Rc And Output  docker1.13 %{VCH-PARAMS} plugin rm test-plugin
    Should Be Equal As Integers  ${rc}  1
    Should Contain  ${output}  does not yet support plugins

Docker plugin set
    ${rc}  ${output}=  Run And Return Rc And Output  docker1.13 %{VCH-PARAMS} plugin set test-plugin test-data
    Should Be Equal As Integers  ${rc}  1
    Should Contain  ${output}  does not yet support plugins
