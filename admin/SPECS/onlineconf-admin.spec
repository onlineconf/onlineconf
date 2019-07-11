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
Requires:       mailru-initd-functions >= 1.11

%description
onlineconf-admin application server. Built from revision %{__revision}.

%prep
%setup -n onlineconf/admin
sed -i 's/\(<link href="[^"]*\.css\|<script src="[^"]*\.js\)"/\1?%{version}"/' static/index.html

%build
%{__rm}   -rf %{_builddir}/onlineconf-admin-build
%{__mkdir} -p %{_builddir}/onlineconf-admin-build/src/gitlab.corp.mail.ru/mydev
%{__cp}    -r %{_builddir}/onlineconf %{_builddir}/onlineconf-admin-build/src/gitlab.corp.mail.ru/mydev/

export GOPATH=%{_builddir}/onlineconf-admin-build
cd %{_builddir}/onlineconf-admin-build/src/gitlab.corp.mail.ru/mydev/onlineconf/admin/go
go build -o %{name} ./

cd ../js
npm run build%{?with_green:-green}

%install
[ "%{buildroot}" != "/" ] && rm -fr %{buildroot}
%{__install} -pD -m0755 %{_builddir}/onlineconf-admin-build/src/gitlab.corp.mail.ru/mydev/onlineconf/admin/go/%{name}  %{buildroot}/%{_localbindir}/%{name}
%{__mkdir} -p %{buildroot}/%{_initrddir} %{buildroot}/%{_localetcdir} %{buildroot}/%{_sysconfdir}/{cron.d,nginx} %{buildroot}/usr/local/www/onlineconf
%{__install} -m 644 etc/%{name}.yaml %{buildroot}/%{_localetcdir}/%{name}.yaml
%{__install} -m 755 init.d/%{name} %{buildroot}/%{_initrddir}/%{name}
%{__cp} -r %{_builddir}/onlineconf-admin-build/src/gitlab.corp.mail.ru/mydev/onlineconf/admin/js/build/* $RPM_BUILD_ROOT/usr/local/www/onlineconf/
%{__cp} -r static $RPM_BUILD_ROOT/usr/local/www/onlineconf/classic
%if %{with green}
sed -i '4s/#FFFFFF/#D6F3D6/; 32s/background: white; //' $RPM_BUILD_ROOT/usr/local/www/onlineconf/classic/css/main.css
%endif
%{__cp} -f etc/nginx.conf $RPM_BUILD_ROOT/etc/nginx/onlineconf.conf
echo "@daily root %{_initrddir}/%{name} remove-old-logs" > %{buildroot}/%{_sysconfdir}/cron.d/%{name}

%files
%defattr(-,root,root,-)
%{_localbindir}/onlineconf-admin
%{_initrddir}/%{name}
%config(noreplace) %{_localetcdir}/%{name}.yaml
%config(noreplace) %{_sysconfdir}/nginx/*
/usr/local/www/onlineconf/*
%{_sysconfdir}/cron.d/%{name}

%post
chkconfig --add %{name}
chkconfig %{name} on

%preun
if [ $1 -eq 0 ]; then
    service %{name} stop > /dev/null
    chkconfig --del %{name}
fi

%changelog
* Mon Mar 19 2012 Aleksey Mashanov <a.mashanov@corp.mail.ru>
- initial version
