package MR::OnlineConf::Admin::Web::Updater;

use Mojo::Base 'MR::OnlineConf::Admin::Web::Controller';

# Internal modules
use MR::OnlineConf::Admin::Storage;

sub mtime {
    my ($self) = @_;
    my $list = MR::OnlineConf::Admin::Storage->select(qq[
        SELECT
            MAX(`MTime`) AS `MTime`
        FROM
            `my_config_tree_log`
    ]);

    $list ||= [{}];

    $self->render(json => $list->[0]);
}

sub config {
    my ($self) = @_;
    my $mtime = $self->param('mtime');
    my $reselect = $self->param('reselect');
    my $list;

    if ($mtime && $reselect) {
        $list = MR::OnlineConf::Admin::Storage->select(qq[
                SELECT
                    t.`ID`, t.`Name`, t.`Path`, l.`Version`, l.`Value`, l.`ContentType`, l.`MTime`, l.`Deleted`
                FROM
                    `my_config_tree_log` l JOIN `my_config_tree` t ON l.`NodeID` = t.`ID`
                WHERE
                    l.`MTime` > LEAST(?, DATE_SUB(NOW(), INTERVAL ? SECOND))
                ORDER BY
                    l.`ID`
            ],
            $mtime,
            $reselect
        );
    } else {
        $list = MR::OnlineConf::Admin::Storage->select(qq[
            SELECT
                `ID`, `Name`, `Path`, `Version`, `Value`, `ContentType`
            FROM
                `my_config_tree`
            WHERE NOT
                `Deleted`
            ORDER BY
                `Path`
        ]);
    }

    $list ||= [];

    $self->render(json => $list);
}

sub activity {
    my ($self) = @_;
    my $host = $self->param('host');
    my $mtime = $self->param('mtime');
    my $version = $self->param('version');

    MR::OnlineConf::Admin::Storage->do(qq[
            REPLACE INTO
                `my_config_activity` (`Host`, `Time`, `Online`, `Package`)
            VALUES
                (?, ?, now(), ?)
        ],
        $host,
        $mtime || 0,
        $version
    );

    $self->render(json => {success => 1});
}

1;
