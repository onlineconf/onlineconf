# Conditionally enable/disable some things in epel7 and up
%if 0%{?rhel} >= 7
%bcond_without systemd
%else
%bcond_with systemd
%endif

%bcond_with green
%define debug_package %{nil}
%define __strip /bin/true

Name:		onlineconf-admin
Version:	%{__version}
Release:	%{__release}%{?dist}

Summary:	onlineconf-admin application server
License:	BSD
Group:		MAILRU
URL:		https://github.com/onlineconf/onlineconf

Source:		https://github.com/onlineconf/onlineconf/archive/%{__revision}/onlineconf-%{__version}.tar.gz

BuildRoot:	%{_tmppath}/%{name}-%{version}-%{release}-buildroot
BuildRequires:	mr-rpm-macros
BuildRequires:	golang
BuildRequires:	golang-bin
BuildRequires:	nodejs
BuildRequires:	rsync
%if 0%{?rhel} >= 9
BuildRequires:	npm
%endif
%if %{with systemd}
BuildRequires:	systemd-devel, systemd-units
Requires:	mailru-systemd-units
%else
Requires:	mailru-initd-functions >= 1.11
%endif


%description
%{name} application server. Built from revision %{__revision}.


%prep

%setup -q -n onlineconf-%{__revision}/admin
%{__sed} -i 's/\(<link href="[^"]*\.css\|<script src="[^"]*\.js\)"/\1?%{version}"/' static/index.html


%build
%if 0%{!?goproxy:1} == 0
export GOPROXY='%{goproxy}'
export GONOSUMDB='*/*'
echo "Set GOPROXY to %{goproxy}, GONOSUMDB to */*"
%else
echo "Set GOPROXY to proxy.golang.org because it is not defined"
export GOPROXY="proxy.golang.org"
%endif
cd go
go build -ldflags="-s -w" -a -gcflags=all=-l -trimpath -o %{name} ./

%if 0%{!?npm_registry:1} == 0
echo "Set npm registry in userconfig to %{npm_registry}"
npm set registry=%{npm_registry}
%endif
cd ../js
npm install
npm run build%{?with_green:-green}


%install
[ "%{buildroot}" != "/" ] && rm -fr %{buildroot}
%{__install} -Dpm 0755 go/%{name}          %{buildroot}%{_localbindir}/%{name}
%{__install} -Dpm 0644 etc/%{name}.yaml    %{buildroot}%{_localetcdir}/%{name}.yaml
%if %{with systemd}
%{__install} -Dpm 0644 etc/%{name}.service %{buildroot}%{_unitdir}/%{name}.service
%else
%{__install} -Dpm 0755 etc/%{name}.init    %{buildroot}%{_initrddir}/%{name}
%endif
%{__mkdir_p}                               %{buildroot}%{_usr}/local/www/onlineconf
rsync -a js/build/                         %{buildroot}%{_usr}/local/www/onlineconf
rsync -a static/                           %{buildroot}%{_usr}/local/www/onlineconf/classic
%if %{with green}
%{__sed} -i '4s/#FFFFFF/#D6F3D6/; 32s/background: white; //' %{buildroot}%{_usr}/local/www/onlineconf/classic/css/main.css
%endif
%{__install} -Dpm 0644 etc/nginx.conf      %{buildroot}%{_sysconfdir}/nginx/onlineconf.conf
%if !%{with systemd}
%{__mkdir_p}                               %{buildroot}%{_sysconfdir}/cron.d
echo "@daily root %{_initrddir}/%{name} remove-old-logs" > %{buildroot}/%{_sysconfdir}/cron.d/%{name}
%endif


%files
%defattr(-,root,root,-)
%{_localbindir}/%{name}
%if %{with systemd}
%{_unitdir}/%{name}.service
%else
%{_initrddir}/%{name}
%endif
%config(noreplace) %{_localetcdir}/%{name}.yaml
%config(noreplace) %{_sysconfdir}/nginx/*
%{_usr}/local/www/onlineconf/*
%if !%{with systemd}
%{_sysconfdir}/cron.d/%{name}
%endif


%post
%if %{with systemd}
echo "Executing systemd post-install tasks"
%if 0%{?systemd_post:1}
    %systemd_post %{name}.service
%else
    /bin/systemctl daemon-reload >/dev/null 2>&1 || :
%endif
%else
echo "Executing System V post-install tasks"
/sbin/chkconfig --add %{name}
/sbin/chkconfig %{name} on
%endif


%preun
%if %{with systemd}
echo "Executing systemd pre-uninstall tasks"
%if 0%{?systemd_preun:1}
    %systemd_preun %{name}.service
%else
    if [ $1 -eq 0 ] ; then
        # Package removal, not upgrade
        /bin/systemctl --no-reload disable %{name}.service > /dev/null 2>&1 || :
        /bin/systemctl stop %{name}.service > /dev/null 2>&1 || :
    fi
%endif
%else
echo "Executing System V pre-uninstall tasks"
if [ $1 -eq 0 ] ; then
    /sbin/service %{name} stop > /dev/null
    /sbin/chkconfig --del %{name}
fi
%endif


%changelog
* Wed Jun 11 2025 Sergei Fedosov <s.fedosov@corp.mail.ru>
- Cleanup and prettify specfile

* Fri Jun 06 2025 Sergei Fedosov <s.fedosov@corp.mail.ru>
- Add support for C9/A9
- Add support for goproxy
- Add support for custom npm registry

* Wed Jan 26 2022 Sergei Fedosov <s.fedosov@corp.mail.ru>
- Add C7 systemd unit

* Mon Mar 19 2012 Aleksey Mashanov <a.mashanov@corp.mail.ru>
- initial version
