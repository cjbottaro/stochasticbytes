require "yaml"
require "hyperwave"

Dir.chdir(__dir__)

hosts = YAML.load_file("hosts.yml")
manager = hosts["managers"].sample

include Hyperwave

each_host(manager) do |host|

  Dir.glob("../secrets/*").each do |secret_path|
    secret_name = File.basename(secret_path)

    host.file(
      desc: "Create temp secret file: #{secret_name}",
      src: secret_path,
      dst: "/tmp/#{secret_name}"
    )

    host.shell(
      desc: "Remove existing secret: #{secret_name}",
      cmd: "docker secret rm #{secret_name}",
      if: "docker secret inspect #{secret_name}"
    )

    host.shell(
      desc: "Create Docker secret: #{secret_name}",
      cmd: "docker secret create #{secret_name} /tmp/#{secret_name}"
    )

    host.shell(
      desc: "Cleanup temp secret file: #{secret_name}",
      cmd: "rm -f /tmp/#{secret_name}"
    )
  end

  Dir.glob("../configs/*").each do |config_path|
    config_name = File.basename(config_path)

    host.file(
      desc: "Create temp config file: #{config_name}",
      src: config_path,
      dst: "/tmp/#{config_name}",
    )

    host.shell(
      desc: "Remove existing config: #{config_name}",
      cmd: "docker config rm #{config_name}",
      if: "docker config inspect #{config_name}"
    )

    host.shell(
      desc: "Create config: #{config_name}",
      cmd: "docker config create #{config_name} /tmp/#{config_name}"
    )

    host.shell(
      desc: "Cleanup: /tmp/#{config_name}",
      cmd: "rm -f /tmp/#{config_name}"
    )
  end

end
