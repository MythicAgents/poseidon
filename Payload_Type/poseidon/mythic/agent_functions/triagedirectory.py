from mythic_payloadtype_container.MythicCommandBase import *
import json


class TriageDirectoryArguments(TaskArguments):
    def __init__(self, command_line):
        super().__init__(command_line)
        self.args = {}

    async def parse_arguments(self):
        if len(self.command_line) == 0:
            self.command_line = "."


class TriageDirectoryCommand(CommandBase):
    cmd = "triagedirectory"
    needs_admin = False
    help_cmd = "triagedirectory [path to directory]"
    description = "Find interesting files within a directory on a host."
    version = 1
    author = "@xorrior"
    argument_class = TriageDirectoryArguments
    attackmapping = ["T1083"]
    browser_script = BrowserScript(script_name="triagedirectory", author="@djhohnstein")

    async def create_tasking(self, task: MythicTask) -> MythicTask:
        return task

    async def process_response(self, response: AgentResponse):
        pass
