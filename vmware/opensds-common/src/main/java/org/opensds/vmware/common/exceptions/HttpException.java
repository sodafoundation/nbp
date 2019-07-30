package org.opensds.vmware.common.exceptions;

public class HttpException extends Exception {
    private long httpResponseCode;

    public HttpException(long httpResponseCode, String message) {
        super(message);
        this.httpResponseCode = httpResponseCode;
    }

    public String toString() {
        return String.format("HTTP exception with code %d: %s", this.httpResponseCode, super.toString());
    }
}
