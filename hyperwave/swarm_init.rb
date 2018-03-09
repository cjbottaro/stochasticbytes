require "hyperwave"
require "yaml"

def chmod_chown_file(host, options)
  host.file(options)

  dst   = options[:dst]
  chmod = options[:chmod]
  chown = options[:chown]

  host.shell(
    desc: options[:desc] + " (chmod #{chmod})",
    cmd: "chmod #{chmod} #{dst}",
    unless: [
      "stat -c %a #{dst}",
      ->(r){ r.stdout == chmod }
    ]
  ) if chmod

  host.shell(
    desc: options[:desc] + " (chown #{chown})",
    cmd: "chown #{chown} #{dst}",
    unless: [
      "stat -c %U.%G #{dst}",
      ->(r){ r.stdout == chown }
    ]
  ) if chown
end

Dir.chdir(__dir__)

hosts = YAML.load_file("hosts.yml")
managers = hosts["managers"]
workers = hosts["workers"]

Hyperwave.each_host(managers + workers) do |host|

  host.file(
    desc: "Add alpine-pkg-glibc repository key",
    src: "https://raw.githubusercontent.com/sgerrand/alpine-pkg-glibc/master/sgerrand.rsa.pub",
    dst: "/etc/apk/keys/sgerrand.rsa.pub",
    unless: "[ -f /etc/apk/keys/sgerrand.rsa.pub ]"
  )

  repo = "-X https://apkproxy.herokuapp.com/sgerrand/alpine-pkg-glibc"

  host.shell(
    desc: "Install glibc",
    cmd: "apk add #{repo} glibc glibc-bin",
    unless: "apk info glibc glibc-bin"
  )

  repos = [
    "-X http://dl-cdn.alpinelinux.org/alpine/edge/main",
    "-X http://dl-cdn.alpinelinux.org/alpine/edge/community"
  ].join(" ")

  host.shell(
    desc: "Install Docker",
    cmd: "apk add #{repos} docker",
    unless: "docker --version"
  )

  host.shell(
    desc: "Enable Docker",
    cmd: "rc-update docker default",
    unless: "rc-status default | grep docker"
  )

  host.shell(
    desc: "Start Docker",
    cmd: "rc-service docker start",
    unless: "rc-service docker status | grep started"
  )

  chmod_chown_file(host,
    desc: "Install docker-compose",
    src: "https://github.com/docker/compose/releases/download/1.19.0/docker-compose-Linux-x86_64",
    dst: "/usr/local/bin/docker-compose",
    chmod: "755",
    chown: "root.docker",
    unless: "[ -f /usr/local/bin/docker-compose ]"
  )

  plugin_name = "cjbottaro/do_storage"
  plugin_alias = "do_storage_v01"
  access_token = File.read("../secrets/do_access_token").chomp

  host.shell(
    desc: "Install volume plugin #{plugin_name} as #{plugin_alias}",
    cmd: <<-LINE,
      docker plugin install #{plugin_name}
      --alias #{plugin_alias}
      --grant-all-permissions
      ACCESS_TOKEN=#{access_token}
    LINE
    unless: "docker plugin inspect #{plugin_alias}"
  )

  # https://stackoverflow.com/questions/27262629/jvm-cant-map-reserved-memory-when-running-in-docker-container/36507784#36507784
  host.shell(
    desc: "For Java to work on Alpine: sysctl -w kernel.pax.softmode=1",
    cmd: "sysctl -w kernel.pax.softmode=1",
    unless: "sysctl kernel.pax.softmode | grep 1"
  )

end

manager_token = nil
worker_token = nil

Hyperwave.each_host(managers.first) do |host|

  host.shell(
    desc: "Init swarm",
    cmd: "docker swarm init",
    if: "docker info | grep 'Swarm: inactive'"
  )

  manager_token = host.shell(
    desc: "Get manager token",
    cmd: "docker swarm join-token manager -q",
    change: false
  ).stdout

  worker_token = host.shell(
    desc: "Get worker token",
    cmd: "docker swarm join-token worker -q",
    change: false
  ).stdout

end

ip_address = managers.first

Hyperwave.each_host(managers) do |host|

  host.shell(
    desc: "Join swarm as manager",
    cmd: "docker swarm join --token #{manager_token} #{ip_address}:2377",
    unless: "docker info | grep 'Swarm: active'"
  )

end

Hyperwave.each_host(workers) do |host|

  host.shell(
    desc: "Join swarm as worker",
    cmd: "docker swarm join --token #{worker_token} #{ip_address}:2377",
    unless: "docker info | grep 'Swarm: active'"
  )

end
