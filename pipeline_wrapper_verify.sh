#!/bin/bash
# This wrapper code is created so that one can run
# code that from the hack directory, otherwise pipeline has issues
# running whereby it cannot find the <workspace>/../hack/common.sh file.
${PIPELINE_WORKSPACE}/hack/verify.sh -v