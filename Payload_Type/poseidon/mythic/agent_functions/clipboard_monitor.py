from mythic_payloadtype_container.MythicCommandBase import *
import json
from mythic_payloadtype_container.MythicRPC import *


class ClipboardMonitorArguments(TaskArguments):
    def __init__(self, command_line, **kwargs):
        super().__init__(command_line, **kwargs)
        self.args = [
            CommandParameter(
                name="duration",
                cli_name="duration",
                display_name="Monitor Duration",
                type=ParameterType.Number,
                description="Number of seconds to monitor the clipboard, or a negative value to do it indefinitely",
                parameter_group_info=[ParameterGroupInfo(ui_position=1)],
            ),
        ]

    async def parse_arguments(self):
        self.load_args_from_json_string(self.command_line)


class ClipboardMonitorCommand(CommandBase):
    cmd = "clipboard_monitor"
    needs_admin = False
    help_cmd = "clipboard_monitor -duration -1"
    description = "Monitor the macOS clipboard for changes every X seconds"
    version = 1
    author = "@its_a_feature_"
    argument_class = ClipboardMonitorArguments
    attackmapping = []
    attributes = CommandAttributes(supported_os=[SupportedOS.MacOS])

    async def create_tasking(self, task: MythicTask) -> MythicTask:
        if task.args.get_arg("duration") < 0:
            task.display_params = "indefinitely"
        else:
            task.display_params = f"for {task.args.get_arg('duration')} seconds"
        return task

    async def process_response(self, response: AgentResponse):
        pass
