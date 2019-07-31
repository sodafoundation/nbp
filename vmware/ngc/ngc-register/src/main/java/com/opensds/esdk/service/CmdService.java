package com.opensds.esdk.service;

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
