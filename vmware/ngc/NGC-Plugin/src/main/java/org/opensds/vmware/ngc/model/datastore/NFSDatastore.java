// Copyright 2019 The OpenSDS Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package org.opensds.vmware.ngc.model.datastore;

public class NFSDatastore extends Datastore{

    private String storageId;

    private String localPath;

    private String remoteHost;

    private String remotePath;

    public String getLocalPath() {
        return localPath;
    }

    public String getRemoteHost() {
        return remoteHost;
    }

    public String getRemotePath() {
        return remotePath;
    }

    public void setLocalPath(String localPath) {
        this.localPath = localPath;
    }

    public void setRemoteHost(String remoteHost) {
        this.remoteHost = remoteHost;
    }

    public void setRemotePath(String remotePath) {
        this.remotePath = remotePath;
    }


    public String getStorageId() {
        return storageId;
    }

    public void setStorageId(String storageId) {
        this.storageId = storageId;
    }

}
