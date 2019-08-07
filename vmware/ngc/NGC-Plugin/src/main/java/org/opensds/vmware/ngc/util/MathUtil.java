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

import java.math.BigDecimal;
import java.math.RoundingMode;
import java.text.NumberFormat;

public class MathUtil
{
    private static final int PRECISION_TWO = 2;

    private static final int PRECISION_ONE = 1;

    private static final int HUNDRED = 100;

    private static final int THREE = 3;

    private MathUtil() {

    }

    /**
     * 根据double提供的精度四舍五入
     *
     * @param number 需要转的数字
     * @param scale 精度
     * @return double 转换后的double
     */
    public static double downScaleToDouble(double number, int... scale) {
        int precision = PRECISION_TWO;
        if (scale.length >= PRECISION_ONE) {
            precision = scale[0];
        }
        BigDecimal bigDec = new BigDecimal(number);
        BigDecimal one = new BigDecimal(PRECISION_ONE);
        BigDecimal newBig = bigDec.divide(one,
                precision,
                BigDecimal.ROUND_HALF_UP);
        return newBig.doubleValue();
    }

    /**
     * 向下转换
     *
     * @param number 数值
     * @param scale 精度
     * @return String 小数位数
     */
    public static String downScaleToString(double number, int... scale) {
        int precision = PRECISION_TWO;
        if (scale.length >= PRECISION_ONE) {
            precision = scale[0];
        }
        BigDecimal bigDec = new BigDecimal(number);
        BigDecimal one = new BigDecimal(PRECISION_ONE);
        BigDecimal newBig = bigDec.divide(one, precision, BigDecimal.ROUND_DOWN);
        return newBig.toString();
    }

    /**
     * Calculate the percentage without a percent sign
     *
     */
    public static String computePercentNoSign(double p1, double p2, int scale) {
        String perCent = computePercent(p1, p2, scale);
        return perCent.replace("%", "");
    }

    /**
     * Calculate the percentage with a percent sign
     *
     */
    public static String computePercent(double p1, double p2, int scale) {
        String str = "";
        if (0 == p2) {
            return "0";
        }
        double p3 = p1 / p2;
        NumberFormat nf = NumberFormat.getPercentInstance();
        nf.setMinimumFractionDigits(scale);
        nf.setMaximumFractionDigits(scale);
        str = nf.format(p3);
        return str;
    }

    /**
     * Calculate the percentage, retain 2 decimal places
     */
    public static String computePercent(double p1, double p2) {
        return computePercent(p1, p2, PRECISION_TWO);
    }

    /**
     * Rounding formatting numbers
     * Non-dot form, default 2 bit precision
     * @param val The value that needs to be converted
     * @param precision The first is the maximum number of decimal places,
     *                  and the second is the minimum number of decimal places.
     * @return String
     */
    public static String parseNumber(Object val, int... precision) {
        return parseNumber(val, false, precision);
    }

    /**
     * Rounding formatting numbers
     * Optional point form, default 2 bit precision
     * @param val The value that needs to be converted
     * @param isGroupingUsed Whether to display points
     * @param precision The first is the maximum number of decimal places,
     *                  and the second is the minimum number of decimal places.
     * @return String
     */
    public static String parseNumber(Object val, Boolean isGroupingUsed,
            int... precision)
    {
        int max = PRECISION_TWO;
        int min = PRECISION_TWO;
        if (precision.length == PRECISION_ONE)
        {
            max = precision[0];
        }
        else if (precision.length == PRECISION_TWO)
        {
            max = precision[0];
            min = precision[1];
        }
        NumberFormat nf = NumberFormat.getNumberInstance();
        //添加舍入模式
        nf.setRoundingMode(RoundingMode.HALF_UP);
        nf.setMaximumFractionDigits(max);
        nf.setMinimumFractionDigits(min > max ? max : min);
        nf.setGroupingUsed(isGroupingUsed);
        return nf.format(val);
    }

    public static Double formatNumber(double p1, double p2)
    {
        if (p2 == 0 || p1 == 0)
        {
            return 0.00;
        }
        String str = "";
        double p3 = p1 / p2;
        double parse = p3 * HUNDRED;
        String doublestr = parse + "";

        int index = doublestr.indexOf(".");
        if (index == -1)
        {
            str = doublestr + ".00";
        }
        else
        {
            str = (doublestr + "0").substring(0, THREE);
        }

        return Double.valueOf(str);
    }

    /**
     * compute log2
     * @param x
     * @return
     */
    public static int get2M(Long x)
    {
        int log_2[] = {0, 1, 2, 2, 3, 3, 3, 3, 4, 4, 4, 4, 4, 4, 4, 4, 5, 5, 5,
                5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 6, 6, 6, 6, 6, 6, 6, 6,
                6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6,
                6, 6, 6, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7,
                7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7,
                7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7,
                7, 7, 7, 7, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8,
                8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8,
                8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8,
                8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8,
                8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8,
                8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8,
                8, 8, 8, 8, 8, 8};
        int l = -1;
        while (x >= 256)
        {
            l += 8;
            x >>= 8;
        }

        return l + log_2[x.intValue()];
    }

    /**
     * Double add
     * @param value1
     * @param value2
     * @return
     */
    public static Double add(Double value1, Double value2) {
        BigDecimal b1 = new BigDecimal(Double.toString(value1));
        BigDecimal b2 = new BigDecimal(Double.toString(value2));
        return b1.add(b2).doubleValue();
    }

    /**
     * double sub
     * @param value1
     * @param value2
     * @return
     */
    public static Double sub(Double value1, Double value2) {
        BigDecimal b1 = new BigDecimal(Double.toString(value1));
        BigDecimal b2 = new BigDecimal(Double.toString(value2));
        return b1.subtract(b2).doubleValue();
    }

    /**
     * double mul
     * @param value1
     * @param value2
     * @return
     */
    public static Double mul(Double value1, Double value2) {
        BigDecimal b1 = new BigDecimal(Double.toString(value1));
        BigDecimal b2 = new BigDecimal(Double.toString(value2));
        return b1.multiply(b2).doubleValue();
    }


    /**
     * double divide
     * @param dividend
     * @param divisor
     * @param scale
     * @return
     */
    public static Double divide(Double dividend, Double divisor, Integer scale) {
        if (scale < 0) {
            throw new IllegalArgumentException("The scale must be a positive integer or zero");
        }
        BigDecimal b1 = new BigDecimal(Double.toString(dividend));
        BigDecimal b2 = new BigDecimal(Double.toString(divisor));
        return b1.divide(b2, scale,RoundingMode.CEILING).doubleValue();
    }

    /**
     * Provides (precise) decimal places for the specified value rounded off.
     *
     * @param value Need rounded number
     * @param scale Keep a few places after the decimal point
     * @return resul
     */
    public static double round(double value,int scale){
        if(scale<0){
            throw new IllegalArgumentException("The scale must be a positive integer or zero");
        }
        BigDecimal b = new BigDecimal(Double.toString(value));
        BigDecimal one = new BigDecimal("1");
        return b.divide(one,scale, RoundingMode.HALF_UP).doubleValue();
    }

}
