###NGC Plugin Register Guide
**Indroduction:** VMware vSphere plugin registration for OpenSDS storage and third party storage.
**Dependencies:**
NGC Plugin module: Module to communicate with vSphere components.
Adapter Manager module: This is the adapter implementation of opensds and thirdparty storage.
PR: https://github.com/opensds/nbp/pull/252

### support https/http
* step 1: Import maven project.
* step 2: Run the maven project and generate the NGC-Register-*.zip package.
* step 3: Unzip the package, and run the script in bin directory.
* step 4: Open the url (https://localhost:8088/homePage or http://localhost:8080/homePage) and register the ngc plugin into vCenter server.
