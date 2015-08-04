Name:           onlineconf-updater2
Version:        %{__version}
Release:        %{__release}%{?dist}

Summary:        onlineconf-updater2 script
License:        BSD
Group:          MAILRU

BuildRoot:      %{_tmppath}/%{name}-%{version}-%{release}-buildroot
BuildArch:      noarch
AutoReq:        0
BuildRequires:  git
BuildRequires:  mr-rpm-macros
Requires:       perl-AnyEvent >= 5.31
Requires:       perl-EV >= 4.02
Requires:       perl-JSON
Requires:       perl-JSON-XS
Requires:       perl-IO-Interface
Requires:       perl-List-MoreUtils
Requires:       perl-Log-Dispatch
Requires:       perl-Mouse
Requires:       perl-MR-DBI >= 20120606.1301
Requires:       perl-Net-IP-CMatch
Requires:       perl-Text-Glob
Requires:       perl-YAML
Requires:       mailru-initd-functions >= 1.11
Conflicts:      perl-MR-Onlineconf < 20120328.1753

%description
onlineconf-updater2 script. Built from revision %{__revision}.

%prep
%setup -n onlineconf/updater2
sed -i "s/our \$VERSION = '1.0';/our \$VERSION = '%{version}';/" lib/MR/OnlineConf/Updater.pm

%build
%__perl Makefile.PL INSTALLDIRS=vendor
%__make %{?_smp_mflags}

%install
[ "%{buildroot}" != "/" ] && rm -fr %{buildroot}
%__make pure_install PERL_INSTALL_ROOT=$RPM_BUILD_ROOT
find $RPM_BUILD_ROOT -type f -name .packlist -exec rm -f {} ';'
find $RPM_BUILD_ROOT -depth -type d -exec rmdir {} 2>/dev/null ';'
%__mkdir -p %{buildroot}/%{_initrddir} %{buildroot}/%{_localetcdir}/onlineconf2 %{buildroot}/%{_sysconfdir}/cron.d
%__install -m 644 etc/onlineconf2.yaml %{buildroot}/%{_localetcdir}/onlineconf2.yaml
%__install -m 755 init.d/onlineconf2 %{buildroot}/%{_initrddir}/onlineconf2
%__mv %{buildroot}/%{_bindir} %{buildroot}/%{_localbindir}
echo "@daily root %{_initrddir}/onlineconf2 remove-old-logs" > %{buildroot}/%{_sysconfdir}/cron.d/%{name}
%_fixperms %{buildroot}/*

%files
%defattr(-,root,root,-)
%{perl_vendorlib}/*
%{_localbindir}/*
%{_initrddir}/onlineconf2
%config(noreplace) %attr(-,update,mail) %{_localetcdir}/onlineconf2.yaml
%dir %attr(755,root,mail) %{_localetcdir}/onlineconf2
%{_sysconfdir}/cron.d/%{name}

%post
chkconfig --add onlineconf2
chkconfig onlineconf2 on

%preun
if [ $1 -eq 0 ]; then
    service onlineconf2 stop > /dev/null
    chkconfig --del onlineconf2
fi

%changelog
* Mon Mar 19 2012 Aleksey Mashanov <a.mashanov@corp.mail.ru>
- initial version
