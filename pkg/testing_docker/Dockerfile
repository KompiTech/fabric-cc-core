FROM couchdb:3.1.1
RUN echo '\n\
[couchdb]\n\
delayed_commits = true\n\
update_lru_on_read = false\n\
[log]\n\
level = error\n\
[cluster]\n\
q=1\n\
n=1\n\
[couch_httpd_auth]\n\
min_iterations = 1\n\
max_iterations = 1\n\
iterations = 1\n\
[httpd]\n\
socket_options = [{nodelay, true}]\n\
[chttpd]\n\
server_options = [{backlog, 128}, {acceptor_pool_size, 16}, {max, 4096}]\n\
' >> /opt/couchdb/etc/local.d/couchdb.ini