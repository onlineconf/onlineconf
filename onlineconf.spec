%define __name perl-MR-Onlineconf
%define __version %(/bin/date +"%Y%m%d.%H%M")
# build type is debug or release
%define __gitrelease master
# don`t strip
%define __strip /bin/true
%define __spec_install_post /usr/lib/rpm/brp-compress

Name:           %{__name}
Version:        %{__version}
Release:        1

Summary:        onlineconf
License:        BSD
Group:          MAIL.RU

BuildRoot:      %{_tmppath}/%{name}-%{version}-%{release}-buildroot
AutoReq:        0
BuildRequires:  git
BuildArch:      noarch
Requires:       perl-AnyEvent
Requires:	    perl-Class-Singleton
Requires:	    perl-DBD-mysql
Requires:	    perl-JSON-XS
Requires:       perl-YAML
Obsoletes:      onlineconf
Provides:       perl-MR-Onlineconf

Source0:	onlineconf.init
Source1:	onlineconf.yaml
Source2:	onlineconf.logrotate

%description
onlineconf

%prep
rm -vfr myutils
git clone gitosis@farm.corp.mail.ru:myutils.git myutils
if [ %{__gitrelease} = "master" ]
then 
	cd myutils && git checkout %{__gitrelease}
	git fetch origin
	git pull origin master
else
	cd myutils && git checkout %{__gitrelease}
fi

%build
cd myutils/onlineconf
%{__perl} Makefile.PL INSTALLDIRS=vendor
make %{?_smp_mflags}

%install
[ "%{buildroot}" != "/" ] && rm -fr %{buildroot}
cd myutils/onlineconf
make pure_install PERL_INSTALL_ROOT=$RPM_BUILD_ROOT
find $RPM_BUILD_ROOT -type f -name .packlist -exec rm -f {} ';'
find $RPM_BUILD_ROOT -depth -type d -exec rmdir {} 2>/dev/null ';'
chmod -R u+w $RPM_BUILD_ROOT/*
mkdir -p %{buildroot}/etc/init.d %{buildroot}/etc/logrotate.d/ %{buildroot}/usr/local/etc/
install -m 755 %SOURCE0 %{buildroot}/etc/init.d/onlineconf
install -m 640 %SOURCE1 %{buildroot}/usr/local/etc/
install -m 644 %SOURCE2 %{buildroot}/etc/logrotate.d/onlineconf

%files
%defattr(-,root,root,-)
/usr/lib/perl5/vendor_perl/*
/usr/local/bin/onlineconf_updater
/etc/init.d/onlineconf
%config(noreplace) %attr(-,root,mail) /usr/local/etc/onlineconf.yaml
%config(noreplace) %{_sysconfdir}/logrotate.d/onlineconf

%post
/sbin/chkconfig --add onlineconf

%changelog
* Wed Jun 22 2011 Eugene Dubravsky <dubravsky@corp.mail.ru>
- Fix package name 

* Tue May 10 2011 Yaroslav Zhavoronkov <zhavoronkov@corp.mail.ru>
- Initial release after porting from logcpd

