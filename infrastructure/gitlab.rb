# These are the only changed settings
external_url 'https://gitlab:10443/'
registry_nginx['enable'] = true
registry_nginx['listen_port'] = 4567
registry_external_url 'https://gitlab:4567'
nginx['ssl_certificate'] = "/etc/gitlab/ssl/gitlab.crt"
nginx['ssl_certificate_key'] = "/etc/gitlab/ssl/gitlab.key"

# Required - otherwise der port from external_url is used - https://docs.gitlab.com/omnibus/settings/nginx.html#setting-the-nginx-listen-port
nginx['listen_port'] = 10443

letsencrypt['enable'] = false
prometheus_monitoring['enable'] = false

# Constrained memory - https://docs.gitlab.com/omnibus/settings/memory_constrained_envs.html
puma['worker_processes'] = 0
sidekiq['max_concurrency'] = 10

gitaly['configuration'] = {
    concurrency: [
      {
        'rpc' => "/gitaly.SmartHTTPService/PostReceivePack",
        'max_per_repo' => 3,
      }, {
        'rpc' => "/gitaly.SSHService/SSHUploadPack",
        'max_per_repo' => 3,
      },
    ],
    # cgroupV2 is still broken - https://forum.gitlab.com/t/gitaly-fails-to-start-unable-to-create-file-in-sys/60455
    #cgroups: {
    #    repositories: {
    #        count: 2,
    #    },
    #    mountpoint: '/sys/fs/cgroup',
    #    hierarchy_root: 'gitaly',
    #    memory_bytes: 500000,
    #    cpu_shares: 512,
    #},
}

gitaly['env'] = {
  'GITALY_COMMAND_SPAWN_MAX_PARALLEL' => '2'
}

gitlab_rails['env'] = {
  'MALLOC_CONF' => 'dirty_decay_ms:1000,muzzy_decay_ms:1000'
}

gitaly['env'] = {
  'MALLOC_CONF' => 'dirty_decay_ms:1000,muzzy_decay_ms:1000'
}