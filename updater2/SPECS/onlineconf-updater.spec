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
Requires:       perl-AnyEvent >= 5.31
Requires:       perl-EV >= 4.02
Requires:       perl-JSON
Requires:       perl-Log-Dispatch
Requires:       perl-Mouse
Requires:       perl-YAML
Requires:       perl-CBOR-XS >= 1.25
Requires:       perl-CDB_File >= 0.98
Requires:       perl-libwww-perl
Requires:       mailru-initd-functions >= 1.11
Conflicts:      perl-MR-Onlineconf < 20120328.1753

%description
onlineconf-updater script. Built from revision %{__revision}.

%prep
%setup -n onlineconf/updater2
sed -i "s/our \$VERSION = '2.0';/our \$VERSION = '%{version}';/" lib/MR/OnlineConf/Updater.pm

%build
%__perl Makefile.PL INSTALLDIRS=vendor
%__make %{?_smp_mflags}

%install
[ "%{buildroot}" != "/" ] && rm -fr %{buildroot}
%__make pure_install PERL_INSTALL_ROOT=$RPM_BUILD_ROOT
find $RPM_BUILD_ROOT -type f -name .packlist -exec rm -f {} ';'
find $RPM_BUILD_ROOT -depth -type d -exec rmdir {} 2>/dev/null ';'
%__mkdir -p %{buildroot}/%{_initrddir} %{buildroot}/%{_localetcdir}/onlineconf %{buildroot}/%{_sysconfdir}/cron.d
%__install -m 755 init.d/onlineconf %{buildroot}/%{_initrddir}/onlineconf
%__mv %{buildroot}/%{_bindir} %{buildroot}/%{_localbindir}
echo "@daily root %{_initrddir}/onlineconf remove-old-logs" > %{buildroot}/%{_sysconfdir}/cron.d/%{name}
%_fixperms %{buildroot}/*

%files
%defattr(-,root,root,-)
%{perl_vendorlib}/*
%{_localbindir}/*
%{_initrddir}/onlineconf
%dir %attr(755,root,mail) %{_localetcdir}/onlineconf
%{_sysconfdir}/cron.d/%{name}

%post
chkconfig --add onlineconf
chkconfig onlineconf on

%preun
if [ $1 -eq 0 ]; then
    service onlineconf stop > /dev/null
    chkconfig --del onlineconf
fi

%changelog
* Mon Mar 19 2012 Aleksey Mashanov <a.mashanov@corp.mail.ru>
- initial version
