import i18next from 'i18next';
import LanguageDetector from 'i18next-browser-languagedetector';
import { initReactI18next } from 'react-i18next';

const resources = {
	en: {
		translation: {
			search: 'Search',
			left: {
				configuration: 'Configuration',
				history: 'History',
				servers: 'Servers',
				access: 'Access',
			},
			button: {
				cancel: 'Cancel',
				close: 'Close',
				load: 'Load',
				ok: 'OK',
				save: 'Save',
			},
			param: {
				path: 'Path',
				summary: 'Summary',
				description: 'Description',
				type: 'Type',
				value: 'Value',
				comment: 'Comment',
				noAccess: 'no access',
				menu: {
					reload: 'Reload',
					view: 'View',
					history: 'History',
					edit: 'Edit',
					describe: 'Describe',
					move: 'Move',
					delete: 'Delete',
					access: 'Access',
					notifications: 'Notifications',
					create: 'Create',
				},
				types: {
					'application/x-null': 'Null',
					'text/plain': 'Text',
					'application/x-template': 'Template',
					'application/json': 'JSON',
					'application/x-yaml': 'YAML',
					'application/x-symlink': 'Symlink',
					'application/x-case': 'Case',
					'application/x-list': 'List',
					'application/x-server': 'ip:port pairs list',
					'application/x-server2': 'Ports for IP list',
				},
				template: {
					hostname: 'full hostname',
					shortHostname: 'short hostname',
					ip: 'host IP address',
					param: 'value of parameter',
				},
				case: {
					by: 'Case by',
					default: 'Default',
					server: 'Server',
					group: 'Group',
					datacenter: 'Datacenter',
					service: 'Service',
				},
				notifications: {
					label: 'Notifications',
					none: 'None',
					noValue: 'Without value',
					withValue: 'With value',
				},
				move: {
					to: 'Move to',
					symlink: 'Leave symlink',
				},
				delete: {
					confirm: 'Are you really want to delete {{param}}?',
				},
				encrypted: {
					flag: 'Encrypted',
					ineligible: 'Null, Symlink and Case values cannot be encrypted; the branches of a Case can.',
					masked: 'encrypted value',
					reveal: 'Reveal',
					hide: 'Hide',
					revealEdit: 'Reveal and edit',
					replace: 'Replace',
					revealHistory: 'Reveal history',
					mask: 'Hide the text',
					unmask: 'Show the text',
					rollbackKeepsEncrypted: 'The parameter is encrypted: the restored value will be stored encrypted too.',
					rollbackNeedsReveal: 'This version is encrypted: reveal the history to restore it into a deleted parameter.',
					declassify: {
						title: 'Store in plaintext?',
						text: 'The value will be stored unencrypted from this version on. Previous versions stay encrypted.',
						confirm: 'Store in plaintext',
					},
				},
			},
			log: {
				author: 'Author',
				branch: 'Branch',
				from: 'From',
				till: 'Till',
				all: 'All',
				load: 'Load',
				deleted: 'deleted',
				rollback: {
					rollback: 'Rollback',
					confirmation: 'Do you want to rollback parameter {{param}} to version {{version}}?',
					current: 'Right now its version is {{version}} with the following value:',
				}
			},
			server: {
				host: 'Server',
				mtime: 'Modification',
				online: 'Online',
				package: 'Version',
				delete: {
					title: 'Delete server?',
					message: 'Delete {{server}} from monitoring?',
				},
			},
			access: {
				createGroup: 'Create group',
				group: 'Group',
				addUser: 'Add user to group "{{group}}"',
				user: 'User',
			},
		},
	},
	ru: {
		translation: {
			search: 'Поиск',
			left: {
				configuration: 'Конфигурация',
				history: 'Журнал',
				servers: 'Серверы',
				access: 'Доступ',
			},
			button: {
				cancel: 'Отменить',
				close: 'Закрыть',
				load: 'Загрузить',
				ok: 'OK',
				save: 'Сохранить',
			},
			param: {
				path: 'Путь',
				summary: 'Резюме',
				description: 'Описание',
				type: 'Тип',
				value: 'Значение',
				comment: 'Комментарий',
				noAccess: 'нет доступа',
				menu: {
					reload: 'Обновить',
					view: 'Посмотреть',
					history: 'Журнал',
					edit: 'Изменить',
					describe: 'Описать',
					move: 'Переместить',
					delete: 'Удалить',
					access: 'Доступ',
					notifications: 'Уведомления',
					create: 'Создать',
				},
				types: {
					'application/x-null': 'Null',
					'text/plain': 'Текст',
					'application/x-template': 'Шаблон',
					'application/json': 'JSON',
					'application/x-yaml': 'YAML',
					'application/x-symlink': 'Symlink',
					'application/x-case': 'Case',
					'application/x-list': 'Список',
					'application/x-server': 'Список пар ip:port',
					'application/x-server2': 'Список портов для ip',
				},
				template: {
					hostname: 'полное имя хоста',
					shortHostname: 'сокращенное имя хоста',
					ip: 'IP-адрес, соответствующий хосту',
					param: 'значение параметра',
				},
				case: {
					by: 'Case по',
					default: 'По умолчанию',
					server: 'Сервер',
					group: 'Группа',
					datacenter: 'Датацентр',
					service: 'Сервис',
				},
				notifications: {
					label: 'Уведомления',
					none: 'Нет',
					noValue: 'Без значения',
					withValue: 'С значением',
				},
				move: {
					to: 'Переместить в',
					symlink: 'Оставить симлинк',
				},
				delete: {
					confirm: 'Вы действительно хотите удалить {{param}}?',
				},
				encrypted: {
					flag: 'Зашифровано',
					ineligible: 'Значения типов Null, Symlink и Case не шифруются; ветки Case — можно.',
					masked: 'зашифрованное значение',
					reveal: 'Показать',
					hide: 'Скрыть',
					revealEdit: 'Показать и изменить',
					replace: 'Заменить',
					revealHistory: 'Показать историю',
					mask: 'Скрыть текст',
					unmask: 'Показать текст',
					rollbackKeepsEncrypted: 'Параметр зашифрован: восстановленное значение тоже будет сохранено зашифрованным.',
					rollbackNeedsReveal: 'Эта версия зашифрована: чтобы восстановить её в удалённый параметр, сначала покажите историю.',
					declassify: {
						title: 'Сохранить открытым текстом?',
						text: 'Начиная с этой версии значение будет храниться незашифрованным. Предыдущие версии останутся зашифрованными.',
						confirm: 'Сохранить открытым',
					},
				},
			},
			log: {
				author: 'Автор',
				branch: 'Ветка',
				from: 'С',
				till: 'По',
				all: 'Все',
				load: 'Загрузить',
				deleted: 'удален',
				rollback: {
					rollback: 'Откатить',
					confirmation: 'Откатить параметр {{param}} на версию {{version}}?',
					current: 'Текущая версия {{version}}, значение:',
				},
			},
			server: {
				host: 'Сервер',
				mtime: 'Модификация',
				online: 'Онлайн',
				package: 'Версия',
				delete: {
					title: 'Удалить сервер?',
					message: 'Удалить {{server}} из мониторинга?',
				},
			},
			access: {
				createGroup: 'Создать группу',
				group: 'Группа',
				addUser: 'Добавить пользователя в группу "{{group}}"',
				user: 'Пользователь',
			},
		},
	},
};

i18next
	.use(initReactI18next)
	.use(LanguageDetector)
	.init({
		resources,
		fallbackLng: 'en',
		interpolation: {
			escapeValue: false,
		},
	});

export default i18next;
