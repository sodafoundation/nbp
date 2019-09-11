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
package org.opensds.vmware.ngc.util;
import org.apache.commons.logging.Log;
import org.apache.commons.logging.LogFactory;
import java.text.ParseException;
import java.text.SimpleDateFormat;
import java.util.Date;
import java.util.TimeZone;

public class TimeUtil
{

    public class SystemTimeZone
    {
        private String timeZone;

        private String timeZoneName;

        private String useDST;// 0-without ,1-use

        private String adjustTime;// dst  0-120min

        private String dateTimeBegin;// dst begin time

        private String dateTimeEnd;// dst end time

        public String getTimeZone()
        {
            return timeZone;
        }

        public void setTimeZone(String timeZone)
        {
            this.timeZone = timeZone;
        }

        public String getTimeZoneName()
        {
            return timeZoneName;
        }

        public void setTimeZoneName(String timeZoneName)
        {
            this.timeZoneName = timeZoneName;
        }

        public String getUseDST()
        {
            return useDST;
        }

        public void setUseDST(String useDST)
        {
            this.useDST = useDST;
        }

        public String getAdjustTime()
        {
            return adjustTime;
        }

        public void setAdjustTime(String adjustTime)
        {
            this.adjustTime = adjustTime;
        }

        public String getDateTimeBegin()
        {
            return dateTimeBegin;
        }

        public void setDateTimeBegin(String dateTimeBegin)
        {
            this.dateTimeBegin = dateTimeBegin;
        }

        public String getDateTimeEnd()
        {
            return dateTimeEnd;
        }

        public void setDateTimeEnd(String dateTimeEnd)
        {
            this.dateTimeEnd = dateTimeEnd;
        }


        @Override
        public String toString() {
            return "SystemTimeZone{" +
                    "timeZone='" + timeZone + '\'' +
                    ", timeZoneName='" + timeZoneName + '\'' +
                    ", useDST='" + useDST + '\'' +
                    ", adjustTime='" + adjustTime + '\'' +
                    ", dateTimeBegin='" + dateTimeBegin + '\'' +
                    ", dateTimeEnd='" + dateTimeEnd + '\'' +
                    '}';
        }
    }

    private static final Log _logger = LogFactory.getLog(TimeUtil.class);

    private static final SimpleDateFormat format1 =  new SimpleDateFormat("yyyy-MM-dd HH:mm:ss");
    private static final SimpleDateFormat resultFormat =  new SimpleDateFormat("yyyy-MM-dd'T'HH:mm:ss'+00:00'");

    static {
       format1.setTimeZone(TimeZone.getTimeZone("GMT"));
       resultFormat.setTimeZone(TimeZone.getTimeZone("GMT"));
    }

    public static String getUTCStringFromFormat1(String source) throws ParseException {
        _logger.debug(String.format("TimeStamp source from Format1 is %s", source));
        Date date = format1.parse(source);
        String result = resultFormat.format(date);
        _logger.debug(String.format("TimeStamp is %s", result));
        return result;
    }

    public static String getUTCStringFromLong(long source){
        _logger.debug(String.format("TimeStamp source from Long is %d", source));
        String result = resultFormat.format(new Date(source));
        _logger.debug(String.format("TimeStamp is %s", result));
        return  result;
    }

    /**
     * convert long time seconds to date format yyyy-MM-dd hh:mm:ss UTC
     * @param timeLong second
     * @return String yyyy-MM-dd hh:mm:ss UTC
     */
    public static String toUTCString(SystemTimeZone systemTimeZone, Long timeLong)
    {
        if (null == systemTimeZone || timeLong == -1)
        {
            return "";
        }
        String timeZone = systemTimeZone.getTimeZone();
        String useDST = systemTimeZone.getUseDST();
        String adjustTime = systemTimeZone.getAdjustTime();
        String dateTimeBegin = systemTimeZone.getDateTimeBegin();
        String dateTimeEnd = systemTimeZone.getDateTimeEnd();
        String result = "";

        Date date = new Date(timeLong * 1000); // 根据long类型的秒数生命一个date类型的时间

        SimpleDateFormat format = new SimpleDateFormat("yyyy-MM-dd'T'HH:mm:ss");
        String timezoneSymbols = timeZone.substring(0, 1);
        String timezoneHour = timeZone.substring(1, 3);
        String timezoneMin = timeZone.substring(4, 6);
        format.setTimeZone(TimeZone.getTimeZone("UTC" + timezoneSymbols + Long.parseLong(timezoneHour)));

        result = format.format(date);

        /** 时区转换 */
        String timeZoneResult = new String();

        if ("1".equals(useDST))
        {
            String formatTimeStr = result.substring(5);
            formatTimeStr = formatTimeStr.replaceAll("[^0-9]", "");
            String begin = dateTimeBegin.replaceAll("[^0-9]", "");
            String end = dateTimeEnd.replaceAll("[^0-9]", "");
            long formatTime = Long.parseLong(formatTimeStr);
            long beginTime = Long.parseLong(begin);
            long endTime = Long.parseLong(end);
            if (formatTime >= beginTime && formatTime <= endTime)
            {
                long offset = Long.parseLong(timezoneHour) * 60 + Long.parseLong(timezoneMin)
                        + ("+".equals(timezoneSymbols) ? 1 : -1) * Long.parseLong(adjustTime);
                timeZoneResult = " UTC" + timezoneSymbols + String.format("%02d", offset / 60) + ":"
                        + String.format("%02d", offset % 60) + " DST";
            }
            else
            {
                timeZoneResult = " UTC" + timeZone;
            }
        }
        else
        {
            timeZoneResult = " UTC" + timeZone;
        }
//        result = result + timeZoneResult;
        return result;
    }
}
