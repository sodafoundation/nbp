package com.opensds.esdk.model;

public class EventsInfo {

    private String eventLocal;

    private String eventName;

    private String eventDescription;

    private String eventTypeId;

    private String eventTypeSchema;

    private String eventSeverity;

    public String getEventName() {
        return eventName;
    }

    public void setEventName(String eventName) {
        this.eventName = eventName;
    }

    public String getEventDescription() {
        return eventDescription;
    }

    public void setEventDescription(String eventDescription) {
        this.eventDescription = eventDescription;
    }

    public String getEventTypeId() {
        return eventTypeId;
    }

    public void setEventTypeId(String eventTypeId) {
        this.eventTypeId = eventTypeId;
    }

    public String getEventTypeSchema() {
        return eventTypeSchema;
    }

    public void setEventTypeSchema(String eventTypeSchema) {
        this.eventTypeSchema = eventTypeSchema;
    }

    public String getEventSeverity() {
        return eventSeverity;
    }

    public void setEventSeverity(String eventSeverity) {
        this.eventSeverity = eventSeverity;
    }

    public String getEventLocal() {
        return eventLocal;
    }

    public void setEventLocal(String eventLocal) {
        this.eventLocal = eventLocal;
    }
}
