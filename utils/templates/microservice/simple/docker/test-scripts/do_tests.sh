#!/usr/bin/env bash

# wait for init service signal, than start test
while [ ! -f /sharedContTest/init_ready ]; do sleep 1; done && \
echo '------*** Start test ***------' && \

# -count=1 don't cache results
# -v verbose


# -- Run single Suite --
# go test -p 1 -run TestRecommendationServiceSuite ../../pkg/recommendation/ -count=1 -v;

# -- Run single test in a suite --
# go test -run TestRecommendationServiceSuite ../../pkg/recommendation -testify.m TestGetDefaultScoresFunctional -count=1 -v;

# -- Run single test --
# go test -run TestGetRecommendedNoAuth ../../pkg/recommendation -count=1 -v;

# -- Run single package test --
#go test -p 1../../pkg/cont/... -count=1 -v;

go test -p 1 ../../pkg/... -count=1 && \

if [[ $DONT_EXIT_ON_END == "yessa" ]]
then
  tail -f /dev/null
fi