# if your proto imports other files
sysl import --format protobufDir \
            --import-paths ./ \
            --output output.sysl \
            --input gateway.proto

# if your proto doesn't import anything you can do this
sysl import --format protobuf \
            --output output.sysl \
            --input gateway.proto
