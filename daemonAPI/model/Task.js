/**
 * Created by tigran on 5/24/15.
 */

module.exports = function(mongoose){
    var schema = new mongoose.Schema({
        task_type: String,
        data: Object,
        cron: { type: Boolean, default: false },
        created: { type: Date, default: Date.now },
        start_time: { type: Date, default: null },
        end_time: { type: Date, default: null },
        done: { type: Boolean, default: false },
        error: { type: Boolean, default: false },
        error_message: String,
        sent: { type: Boolean, default: false },
        sent_time: { type: Date, default: null },
        done_time: { type: Date, default: null },
        daemon: { type: mongoose.Schema.Types.ObjectId, ref: 'Daemon' }
    });

    schema.methods.taskInfo = function(){
        var $this = this;
        return {
            task_id: $this._id.toString(),
            start_time: $this.sent_time,
            end_time: $this.done_time,
            error: $this.error,
            message: $this.error_message,
            done: $this.done
        };
    };

    return schema;
};