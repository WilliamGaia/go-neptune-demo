#!/bin/bash
curl -v POST https://pkm-demo-instance-real.cluster-cfhqexhrq16e.ap-northeast-1.neptune.amazonaws.com:8182/loader \
-H 'Content-Type: application/json' \
-d '{"source":"s3://pkm-neptune-demo/bulk_raw_real/vertices",
"format":"opencypher", 
"iamRoleArn":"arn:aws:iam::383973027857:role/s3_reader", 
"region":"ap-northeast-1", 
"failOnError":"FALSE", 
"parallelism":"MEDIUM", 
"userProvidedEdgeIds":"TRUE"}' -w "@performance-format.txt" -o insert_vertices_interval.txt
