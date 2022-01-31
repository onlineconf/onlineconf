# Conditionally enable/disable some things in epel7 and up
%if 0%{?rhel} >= 7
%bcond_without systemd
%else
%bcond_with systemd
%endif

%bcond_with green
%define debug_package %{nil}
%define __strip /bin/true

Name:           onlineconf-admin
Version:        %{__version}
Release:        %{__release}%{?dist}

Summary:        onlineconf-admin application server
License:        BSD
Group:          MAILRU

BuildRoot:      %{_tmppath}/%{name}-%{version}-%{release}-buildroot
BuildRequires:  mr-rpm-macros
BuildRequires:  perl(ExtUtils::MakeMaker)
BuildRequires:  golang
BuildRequires:  golang-bin
BuildRequires:  nodejs
%if %{with systemd}
BuildRequires:  systemd-devel, systemd-units
Requires:       mailru-systemd-units
%else
Requires:       mailru-initd-functions >= 1.11
%endif


%description
onlineconf-admin application server. Built from revision %{__revision}.


%prep
%setup -q -c -n %{name}-%{version}
%setup -T -D -n %{name}-%{version}/onlineconf/admin
sed -i 's/\(<link href="[^"]*\.css\|<script src="[^"]*\.js\)"/\1?%{version}"/' static/index.html


%build
cd go
go build -mod=vendor -o %{name} ./

cd ../js
npm run build%{?with_green:-green}


%install
[ "%{buildroot}" != "/" ] && rm -fr %{buildroot}
%{__install} -pD -m0755 go/%{name}  %{buildroot}/%{_localbindir}/%{name}
%{__mkdir} -p %{buildroot}/%{_initrddir} %{buildroot}/%{_localetcdir} %{buildroot}/%{_sysconfdir}/{cron.d,nginx} %{buildroot}/usr/local/www/onlineconf
%{__install} -m 644 etc/%{name}.yaml %{buildroot}/%{_localetcdir}/%{name}.yaml
%if %{with systemd}
%{__install} -pD -m 644 etc/%{name}.service %{buildroot}%{_unitdir}/%{name}.service
%else
%{__install} -pD -m 755 etc/%{name}.init %{buildroot}%{_initrddir}/%{name}
%endif
%{__cp} -r js/build/* $RPM_BUILD_ROOT/usr/local/www/onlineconf/
%{__cp} -r static $RPM_BUILD_ROOT/usr/local/www/onlineconf/classic
%if %{with green}
sed -i '4s/#FFFFFF/#D6F3D6/; 32s/background: white; //' $RPM_BUILD_ROOT/usr/local/www/onlineconf/classic/css/main.css
%endif
%{__cp} -f etc/nginx.conf $RPM_BUILD_ROOT/etc/nginx/onlineconf.conf
%if !%{with systemd}
echo "@daily root %{_initrddir}/%{name} remove-old-logs" > %{buildroot}/%{_sysconfdir}/cron.d/%{name}
%endif


%files
%defattr(-,root,root,-)
%{_localbindir}/%{name}
%if %{with systemd}
%config %{_unitdir}/%{name}.service
%else
%{_initrddir}/%{name}
%endif
%config(noreplace) %{_localetcdir}/%{name}.yaml
%config(noreplace) %{_sysconfdir}/nginx/*
/usr/local/www/onlineconf/*
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
* Wed Jan 26 2022 Sergei Fedosov <s.fedosov@corp.mail.ru>
- Add C7 systemd unit

* Mon Mar 19 2012 Aleksey Mashanov <a.mashanov@corp.mail.ru>
- initial version
