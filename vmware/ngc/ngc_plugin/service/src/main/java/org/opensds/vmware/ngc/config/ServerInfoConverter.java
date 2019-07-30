package org.opensds.vmware.ngc.config;

import com.vmware.vise.usersession.ServerInfo;
import com.vmware.vise.usersession.UserSession;
import com.vmware.vise.usersession.UserSessionService;
import org.springframework.core.convert.converter.Converter;


public class ServerInfoConverter implements Converter<String, ServerInfo> {
    private UserSessionService userSessionService;

    public ServerInfoConverter(UserSessionService userSessionService) {
        this.userSessionService = userSessionService;
    }

    @Override
    public ServerInfo convert(String serviceGuid) {
        UserSession userSession = userSessionService.getUserSession();
        if (userSession == null) {
            throw new RuntimeException("userSession is null");
        }
        for (ServerInfo serverInfo : userSession.serversInfo) {
            if (serverInfo.serviceGuid.equalsIgnoreCase(serviceGuid)) {
                return serverInfo;
            }
        }
        throw new RuntimeException("Can not find serverInfo");
    }
}
