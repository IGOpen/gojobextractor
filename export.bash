#! /bin/bash

cp config.gcfg config.gcfg.example

tar zcf -  --exclude '*.test' config.gcfg.example coverletter.csv mazurov_cv.pdf usefull.bash bin/ templates/ |ssh juno@104.236.237.125 tar zxf - -C gojobextractor

#scp juno@104.236.237.125:gojobextractor/mytags.csv .


