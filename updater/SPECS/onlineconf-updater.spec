# Conditionally enable/disable some things in epel7 and up
%if 0%{?rhel} >= 7
%bcond_without systemd
%else
%bcond_with systemd
%endif

# do not create debuginfo packages
%define _enable_debug_packages 0
%define debug_package %{nil}
%define __strip /bin/true

Name:		onlineconf-updater
Version:	%{__version}
Release:	%{__release}%{?dist}
Summary:	GoLang flavour of onlineconf-updater
License:	BSD
Group:		MAILRU
URL:		https://github.com/onlineconf/onlineconf

Source:		https://github.com/onlineconf/onlineconf/archive/%{__revision}/onlineconf-%{version}.tar.gz

BuildRoot:	%{_tmppath}/%{name}-%{version}-%{release}-buildroot
BuildRequires:	ca-certificates
BuildRequires:	golang
BuildRequires:	git
BuildRequires:	mr-rpm-macros
%if %{with systemd}
BuildRequires:	systemd-devel, systemd-units
%endif

%if %{with systemd}
Requires:	mailru-systemd-units
%else
Requires:	mailru-initd-functions >= 1.11
%endif
Conflicts:	perl-MR-Onlineconf < 20120328.1753

%description
GoLang flavour of onlineconf-updater. Built from revision %{__revision}.

%prep
%setup -q -n onlineconf-%{__revision}/updater

%build
%if 0%{!?goproxy:1} == 0
export GOPROXY='%{goproxy}'
export GONOSUMDB='*/*'
echo "Set GOPROXY to %{goproxy}, GONOSUMDB to */*"
%else
echo "Set GOPROXY to proxy.golang.org because it is not defined"
export GOPROXY="proxy.golang.org"
%endif
# Set proper version of app
%{__sed} -i 's|const version = ".*"|const version = "%{version}"|' updater/version.go
go build -ldflags='-s -w' -a -gcflags=all=-l -trimpath -o %{name} ./


%install
%{__mkdir_p}                                  %{buildroot}%{_localetcdir}/onlineconf

%if %{with systemd}
%{__install} -Dpm 0644 etc/onlineconf.service %{buildroot}%{_unitdir}/onlineconf.service
%else
%{__install} -Dpm 0755 etc/onlineconf.init    %{buildroot}%{_initrddir}/onlineconf
%endif

%{__install} -Dpm 0755 onlineconf-updater     %{buildroot}%{_localbindir}/%{name}

%if !%{with systemd}
%{__mkdir_p}                                  %{buildroot}%{_sysconfdir}/cron.d
echo "@daily root %{_initrddir}/onlineconf remove-old-logs" > %{buildroot}/%{_sysconfdir}/cron.d/%{name}
%endif



%post
%if %{with systemd}
echo "Executing systemd post-install tasks"
%if 0%{?systemd_post:1}
    %systemd_post onlineconf.service
%else
    /bin/systemctl daemon-reload >/dev/null 2>&1 || :
%endif
%else
echo "Executing System V post-install tasks"
/sbin/chkconfig --add onlineconf
/sbin/chkconfig onlineconf on
%endif


%preun
%if %{with systemd}
echo "Executing systemd pre-uninstall tasks"
%if 0%{?systemd_preun:1}
    %systemd_preun onlineconf.service
%else
    if [ $1 -eq 0 ] ; then
        # Package removal, not upgrade
        /bin/systemctl --no-reload disable onlineconf.service > /dev/null 2>&1 || :
        /bin/systemctl stop onlineconf.service > /dev/null 2>&1 || :
    fi
%endif
%else
echo "Executing System V pre-uninstall tasks"
if [ $1 -eq 0 ] ; then
    /sbin/service onlineconf stop > /dev/null
    /sbin/chkconfig --del onlineconf
fi
%endif


%files
%defattr(-,root,root,-)
%{_localbindir}/*

%if %{with systemd}
%config %{_unitdir}/onlineconf.service
%else
%{_initrddir}/onlineconf
%endif

%dir %attr(755,root,mail) %{_localetcdir}/onlineconf

%if !%{with systemd}
%{_sysconfdir}/cron.d/%{name}
%endif


%changelog
* Wed Jun 11 2025 Sergei Fedosov <s.fedosov@corp.mail.ru>
- Introduce support for Almalinux-9
- Cleanup and fixup specfile

* Wed Apr 24 2019 Sergei Fedosov <s.fedosov@corp.mail.ru>
- Revamp spec to build onlineconf-updater-go

* Tue Jun 13 2017 Evgenii Molchanov <e.molchanov@corp.mail.ru>
- Add C7 systemd unit

* Mon Mar 19 2012 Aleksey Mashanov <a.mashanov@corp.mail.ru>
- initial version
