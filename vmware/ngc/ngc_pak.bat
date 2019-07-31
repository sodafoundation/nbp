@echo on 
echo step one....
cd ngc-plugin/makefile/
call ant -buildfile ngc_part0.xml
echo step two....
cd ../../ngc-register/
mvn package