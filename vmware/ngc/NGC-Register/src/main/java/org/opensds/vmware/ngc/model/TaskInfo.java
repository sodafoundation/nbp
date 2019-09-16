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

package org.opensds.vmware.ngc.model;

public class TaskInfo {

    private String taskId;

    private String taskLabel;

    private String taskLabelValue;

    private String taskSummary;

    private String taskSummaryValue;

    private String tasklocal;

    public String getTaskId() {
        return taskId;
    }

    public void setTaskId(String taskId) {
        this.taskId = taskId;
    }


    public String getTaskLabel() {
        return taskLabel;
    }

    public void setTaskLabel(String taskLabel) {
        this.taskLabel = taskLabel;
    }


    public String getTaskLabelValue() {
        return taskLabelValue;
    }


    public void setTaskLabelValue(String taskLabelValue) {
        this.taskLabelValue = taskLabelValue;
    }

    public String getTaskSummary() {
        return taskSummary;
    }


    public void setTaskSummary(String taskSummary) {
        this.taskSummary = taskSummary;
    }


    public String getTaskSummaryValue() {
        return taskSummaryValue;
    }


    public void setTaskSummaryValue(String taskSummaryValue) {
        this.taskSummaryValue = taskSummaryValue;
    }


    public void setTasklocal(String tasklocal) {
        this.tasklocal = tasklocal;
    }


    public String getTasklocal() {
        return tasklocal;
    }
}
