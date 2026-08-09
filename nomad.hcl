data_dir  = "/mnt/c/Users/eurek/Downloads/services" #"C:/Users/eurek/Downloads/services/nomad_data"

ui {
  enabled =  true
}

bind_addr = "0.0.0.0" # the default

advertise {
  # Defaults to the first private IP address.
  http = "192.168.1.1"
  rpc  = "192.168.1.1"
  serf = "192.168.1.1:5648" # non-default ports may be specified
}

server {
  enabled          = true
  bootstrap_expect = 1
}

client {
  enabled       = true
}

plugin "raw_exec" {
  config {
    enabled = true
  }
}

#consul {
#  address = "1.2.3.4:8500"
#}

