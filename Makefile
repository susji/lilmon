#
# This Makefile is GNU-style, and the lack of uppercase `PREFIX` may surprise
# BSD-style build environments.
#
SRC := config.go db.go graph.go main.go measure.go metrics.go \
       protect.go protect_openbsd.go serve.go settings.go types.go

GO ?= go

INSTALL ?= install
INSTALL_DATA ?= $(INSTALL) -m 0644
MKDIR_P ?= mkdir -p

DESTDIR ?=
prefix ?= /usr/local
exec_prefix ?= $(prefix)
bindir ?= $(exec_prefix)/bin
sysconfdir ?= /etc
localstatedir ?= /var

.PHONY: all
all: lilmon

lilmon: $(SRC)
	$(GO) build -o lilmon .

.PHONY: clean
clean:
	rm -f lilmon

.PHONY: check
check: test

.PHONY: test
test:
	$(GO) test

.PHONY: install
install: lilmon lilmon.ini.example lilmon.template.example
	$(MKDIR_P) $(DESTDIR)$(bindir)
	$(MKDIR_P) $(DESTDIR)$(sysconfdir)/lilmon
	$(MKDIR_P) $(DESTDIR)$(localstatedir)/lilmon/db
	$(INSTALL) -m 0755 lilmon $(DESTDIR)$(bindir)
	$(INSTALL_DATA) lilmon.ini.example $(DESTDIR)$(sysconfdir)/lilmon.ini
	$(INSTALL_DATA) lilmon.template.example $(DESTDIR)$(sysconfdir)/lilmon.template
	@echo 'You probably want to chown the directory'
	@echo
	@echo '  $(DESTDIR)$(localstatedir)/lilmon/db'
	@echo
	@echo 'to the non-privileged user who runs the daemons.'
	@echo
	@echo 'Edit the files under'
	@echo
	@echo	'  $(DESTDIR)$(sysconfdir)/lilmon'
	@echo
	@echo 'to your liking before starting the program. If the lilmon database does not'
	@echo 'exist, start the `measure` mode first as it will create and migrate the'
	@echo 'database if it does not exist.'
