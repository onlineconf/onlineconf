Name:           perl-MR-OnlineConf
Version:        %{__version}
Release:        %{__release}%{?dist}
Summary:        onlineconf perl client
License:        MAILRU
Group:          MAILRU

AutoReq:        0
BuildArch:      noarch
BuildRoot:      %{_tmppath}/%{name}-%{?version}-%{?release}-buildroot

Requires:       perl(:MODULE_COMPAT_%(eval "`%{__perl} -V:version`"; echo $version))
Requires:       perl-Class-Singleton
Requires:       perl-JSON-XS
Requires:       perl-CBOR-XS >= 1.25
Requires:       perl-YAML
Requires:       perl-CDB_File >= 0.98
Requires:       onlineconf-updater >= 20160208.2139
Provides:       perl-MR-Onlineconf = %{__version}-%{__revision}%{?dist}
Obsoletes:      onlineconf
Obsoletes:      perl-MR-Onlineconf

%description
Onlineconf perl client. Built from revision %{__revision}.

%prep
%setup -n onlineconf/client

%build
%__perl Makefile.PL INSTALLDIRS=vendor
%__make %{?_smp_mflags}

%install
[ "%{buildroot}" != "/" ] && rm -fr %{buildroot}
%__make pure_install PERL_INSTALL_ROOT=$RPM_BUILD_ROOT
find $RPM_BUILD_ROOT -type f -name .packlist -exec rm -f {} ';'
find $RPM_BUILD_ROOT -depth -type d -exec rmdir {} 2>/dev/null ';'
chmod -R u+w $RPM_BUILD_ROOT/*
%__mkdir -p $RPM_BUILD_ROOT%{_tmppath}
touch $RPM_BUILD_ROOT%{_tmppath}/onlineconf_error.txt
%_fixperms $RPM_BUILD_ROOT/*

%files
%defattr(-,root,root,-)
%{perl_vendorlib}/*
%attr(664,mail,mail) %{_tmppath}/onlineconf_error.txt

%changelog
* Mon Mar 19 2012 Aleksey Mashanov <a.mashanov@corp.mail.ru>
- move updater to separate package
* Wed Jul 04 2011 Sergey Panteleev <panteleev@corp.mail.ru>
- chown onlineconf.yaml to update:mail, moved to noarch repo
* Wed Jun 22 2011 Eugene Dubravsky <dubravsky@corp.mail.ru>
- Fix package name
* Tue May 10 2011 Yaroslav Zhavoronkov <zhavoronkov@corp.mail.ru>
- Initial release after porting from logcpd
