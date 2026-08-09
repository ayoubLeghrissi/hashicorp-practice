job "frontend" {
  datacenters = ["dc1"]

  group "front" {
    network {
      port "http" {
        static = "3000"
      }
    }
    task "server" {
      driver = "docker"

      config {
        image = "services-frontend:latest"
        ports = ["http"]
       
      }
    }
  }
}
