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

package org.opensds.vmware.ngc.service;

import org.apache.logging.log4j.LogManager;
import org.apache.logging.log4j.Logger;
import org.springframework.boot.CommandLineRunner;
import org.springframework.stereotype.Component;


@Component
public class CmdService implements CommandLineRunner {

    private static final Logger logger = LogManager.getLogger(CommandLineRunner.class);

    @Override
    public void run(String... args) throws Exception {
       /* Options options = new Options();
        options.addOption("h", false, "list help");
        options.addOption("t", true, "set time on system");
        try {
            CommandLineParser parser = new DefaultParser();
            CommandLine cmd = parser.parse(options, args);


        }catch (ParseException ex) {
            logger.error("cmd error: " + ex.getMessage());
        }*/
    }
}
