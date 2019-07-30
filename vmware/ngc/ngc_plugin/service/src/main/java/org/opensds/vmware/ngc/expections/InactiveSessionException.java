package org.opensds.vmware.ngc.expections;


public class InactiveSessionException extends Exception{

    public InactiveSessionException() {
        super("The session is inactive or object not connected.");
    }

    public InactiveSessionException(String message) {
        super(message);
    }

    public InactiveSessionException(String message, Throwable cause) {
        super(message, cause);
    }

    public InactiveSessionException(Throwable cause) {
        super(cause);
    }

    public InactiveSessionException(String message, Throwable cause, boolean enableSuppression, boolean writableStackTrace) {
        super(message, cause, enableSuppression, writableStackTrace);
    }
}
