# Conditionally enable/disable some things in epel7
%if 0%{?rhel} == 7
%bcond_without systemd
%endif

Name:           onlineconf-updater
Version:        %{__version}
Release:        %{__release}%{?dist}

Summary:        onlineconf-updater script
License:        BSD
Group:          MAILRU

BuildRoot:      %{_tmppath}/%{name}-%{version}-%{release}-buildroot
BuildArch:      noarch
AutoReq:        0
BuildRequires:  git
BuildRequires:  mr-rpm-macros
BuildRequires:  perl(ExtUtils::MakeMaker)
%if %{with systemd}
BuildRequires:  systemd-devel, systemd-units
%endif
Requires:       perl-AnyEvent >= 5.31
Requires:       perl-EV >= 4.02
Requires:       perl-JSON
Requires:       perl-Log-Dispatch
Requires:       perl-Mouse
Requires:       perl-YAML
Requires:       perl-CBOR-XS >= 1.25
Requires:       perl-CDB_File >= 0.98
Requires:       perl-libwww-perl
%if %{with systemd}
Requires:       mailru-systemd-units
%else
Requires:       mailru-initd-functions >= 1.11
%endif
Conflicts:      perl-MR-Onlineconf < 20120328.1753

%description
onlineconf-updater script. Built from revision %{__revision}.

%prep
%setup -n onlineconf/updater
sed -i "s/our \$VERSION = '2.0';/our \$VERSION = '%{version}';/" lib/MR/OnlineConf/Updater.pm

%build
%__perl Makefile.PL INSTALLDIRS=vendor
%__make %{?_smp_mflags}

%install
[ "%{buildroot}" != "/" ] && rm -fr %{buildroot}
%{__make} pure_install PERL_INSTALL_ROOT=$RPM_BUILD_ROOT

find $RPM_BUILD_ROOT -type f -name .packlist -exec rm -f {} ';'
find $RPM_BUILD_ROOT -depth -type d -exec rmdir {} 2>/dev/null ';'

%{__mkdir} -p %{buildroot}%{_localetcdir}/onlineconf
%{__mkdir} -p %{buildroot}%{_sysconfdir}/cron.d
%if %{with systemd}
%{__mkdir} -p %{buildroot}%{_unitdir}
%else
%{__mkdir} -p %{buildroot}%{_initrddir}
%endif

%if %{with systemd}
%{__install} -m 644 etc/onlineconf.service %{buildroot}%{_unitdir}/onlineconf.service
%else
%{__install} -m 755 etc/onlineconf.init %{buildroot}%{_initrddir}/onlineconf
%endif

%{__mv} %{buildroot}/%{_bindir} %{buildroot}%{_localbindir}
echo "@daily root %{_initrddir}/onlineconf remove-old-logs" > %{buildroot}/%{_sysconfdir}/cron.d/%{name}
%_fixperms %{buildroot}/*

%files
%defattr(-,root,root,-)
%{perl_vendorlib}/*
%{_localbindir}/*
%if %{with systemd}
%config %{_unitdir}/onlineconf.service
%else
%{_initrddir}/onlineconf
%endif

%dir %attr(755,root,mail) %{_localetcdir}/onlineconf
%{_sysconfdir}/cron.d/%{name}

%post
%if %{with systemd}
echo "Executing systemd post-install tasks"
%if 0%{?systemd_post:1}
    %systemd_post onlineconf.service
%else
    if [ $1 -eq 1 ] ; then
        # Initial installation
        /bin/systemctl daemon-reload >/dev/null 2>&1 || :
    fi
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

%changelog
* Tue Jun 13 2017 Evgenii Molchanov <e.molchanov@corp.mail.ru>
- Add C7 systemd unit
* Mon Mar 19 2012 Aleksey Mashanov <a.mashanov@corp.mail.ru>
- initial version
