# RFetch
## Get remote execution results and save them (designed as part of monitoring-dash)


### Configuration file notes:

1. JSON supports "key":"value" only on one string, so PEM-encoded private key need convertion to one-line encoded PEM. To achieve it with AWK use this:
```awk 'NF {sub(/\r/, ""); printf "%s\\n",$0;}' decrypted-key.pem```
2. To test WMI query use standart Windows utility ```wbemtest.exe```, default namespace: ```root\CIMV2```
