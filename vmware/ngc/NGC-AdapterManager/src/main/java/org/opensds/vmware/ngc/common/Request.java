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

package org.opensds.vmware.ngc.common;

import java.security.SecureRandom;
import java.security.cert.X509Certificate;
import java.util.HashMap;
import java.util.Map;

import javax.net.ssl.HostnameVerifier;
import javax.net.ssl.SSLContext;
import javax.net.ssl.SSLSession;
import javax.net.ssl.TrustManager;
import javax.net.ssl.X509TrustManager;

import org.apache.http.client.methods.HttpDelete;
import org.apache.http.client.methods.HttpGet;
import org.apache.http.client.methods.HttpEntityEnclosingRequestBase;
import org.apache.http.client.methods.HttpPost;
import org.apache.http.client.methods.HttpPut;
import org.apache.http.client.methods.HttpUriRequest;
import org.apache.http.HttpResponse;
import org.apache.http.impl.client.HttpClientBuilder;
import org.apache.http.impl.client.CloseableHttpClient;
import org.apache.http.util.EntityUtils;

import org.opensds.vmware.ngc.exceptions.HttpException;

public class Request {
    public interface RequestHandler {
        void setRequestBody(HttpEntityEnclosingRequestBase req, Object body);
        Object parseResponseBody(String body);
    }

    protected String ip;
    protected int port;
    protected String url;
    protected CloseableHttpClient client;
    protected RequestHandler handler;
    protected HashMap<String, String> headers;

    public Request(String ip, int port, RequestHandler handler) throws Exception {
        this.ip = ip;
        this.port = port;
        this.url = String.format("https://%s:%d", ip, port);

        HttpClientBuilder httpClientBuilder = HttpClientBuilder.create();
        httpClientBuilder.setSSLHostnameVerifier(new HostnameVerifier() {
            public boolean verify(String hostname, SSLSession session) {return true;}
        });

        TrustManager[] cert = new TrustManager[] {new X509TrustManager() {
            public X509Certificate[] getAcceptedIssuers() {return null;}
            public void checkClientTrusted(X509Certificate[] certs, String authType) {}
            public void checkServerTrusted(X509Certificate[] certs, String authType) {}
        }
        };

        SSLContext sc = SSLContext.getInstance("SSL");
        sc.init(null, cert, new SecureRandom());
        httpClientBuilder.setSSLContext(sc);

        this.client = httpClientBuilder.build();
        this.handler = handler;
        this.headers = new HashMap<>();
    }

    public void close() throws Exception {
        this.client.close();
    }

    public void setUrl(String url) {
        this.url = url;
    }

    public void setHeaders(String key, String value) {
        this.headers.put(key, value);
    }

    Object call(HttpUriRequest req) throws Exception {
        for(Map.Entry<String, String> entry: this.headers.entrySet()) {
            String k = entry.getKey();
            String v = entry.getValue();

            req.setHeader(k, v);
        }

        HttpResponse response = this.client.execute(req);
        int responseCode = response.getStatusLine().getStatusCode();

        if (responseCode >= 300) {
            String reason = response.getStatusLine().getReasonPhrase();
            throw new HttpException(responseCode, reason);
        }

        String result = EntityUtils.toString(response.getEntity(), "utf-8");
        return this.handler.parseResponseBody(result);
    }

    public Object get(String url) throws Exception {
        HttpGet httpGet = new HttpGet(this.url + url);
        return call(httpGet);
    }

    public Object post(String url, Object body) throws Exception {
        HttpPost httpPost = new HttpPost(this.url + url);
        this.handler.setRequestBody(httpPost, body);
        return call(httpPost);
    }

    public Object put(String url, Object body) throws Exception {
        HttpPut httpPut = new HttpPut(this.url + url);
        this.handler.setRequestBody(httpPut, body);
        return call(httpPut);
    }

    public Object delete(String url) throws Exception {
        HttpDelete httpDelete = new HttpDelete(this.url + url);
        return call(httpDelete);
    }
}
