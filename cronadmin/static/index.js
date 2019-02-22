new Vue({
    el: '#app',
    data: function () {
        return {
            activeName: 'second',
            menus: [
                {name: 'Tasks', active: true},
                {name: 'Nodes', active: false},
            ],
            loading: true,
            tableData: [],

            dialogFormVisible: false,
            formTask: {
                name: '',
                cmd: '',
                cron_line: '',
            },
            formLabelWidth: '120px'
        }
    },
    created: function () {
        this.getList()
    },
    methods: {
        changeMenu: function (menu) {
            this.menus.forEach(function (menu) {
                menu.active = false
            })
            menu.active = true
        },
        getList: function () {
            axios.get('/api/tasks')
                .then((resp) => {
                    this.tableData = resp.data
                    this.loading = false
                })
        },
        addTask: function () {
            this.dialogFormVisible = false
            axios.post('/api/task/save', this.formTask)
                .then((resp) => {
                    this.getList()
                })
        },
        delTask: function (task) {
            axios.post('/api/task/del', task)
                .then((resp) => {
                    this.getList()
                })
        }
    }
});