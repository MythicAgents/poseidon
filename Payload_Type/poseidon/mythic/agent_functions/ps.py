from mythic_payloadtype_container.MythicCommandBase import *
import json


class PsArguments(TaskArguments):
    def __init__(self, command_line):
        super().__init__(command_line)
        self.args = {}

    async def parse_arguments(self):
        self.add_arg("regex_filter", self.command_line)


class PsCommand(CommandBase):
    cmd = "ps"
    needs_admin = False
    help_cmd = "ps [regex name matching]"
    description = "Get a process listing"
    version = 1
    supported_ui_features = ["process_browser:list"]
    author = "@xorrior, @djhohnstein, @its_a_feature"
    argument_class = PsArguments
    attackmapping = ["T1057"]
    browser_script = BrowserScript(script_name="ps", author="@djhohnstein")

    async def create_tasking(self, task: MythicTask) -> MythicTask:
        task.display_params = task.args.command_line
        return task

    async def process_response(self, response: AgentResponse):
        pass
