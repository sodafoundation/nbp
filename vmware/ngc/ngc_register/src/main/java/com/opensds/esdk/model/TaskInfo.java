package com.opensds.esdk.model;

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
