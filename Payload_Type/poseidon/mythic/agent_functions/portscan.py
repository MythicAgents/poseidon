from mythic_payloadtype_container.MythicCommandBase import *
import json


class PortScanArguments(TaskArguments):
    def __init__(self, command_line, **kwargs):
        super().__init__(command_line, **kwargs)
        self.args = [
            CommandParameter(
                name="ports",
                type=ParameterType.String,
                description="List of ports to scan. Can use the dash separator to specify a range.",
                parameter_group_info=[
                    ParameterGroupInfo(
                        ui_position=2
                    )
                ]
            ),
            CommandParameter(
                name="hosts",
                type=ParameterType.Array,
                description="List of host IPs or CIDR notations",
                parameter_group_info=[
                    ParameterGroupInfo(
                        ui_position=1
                    )
                ]
            ),
        ]

    async def parse_arguments(self):
        if len(self.command_line) == 0:
            raise Exception("Must supply arguments")
        else:
            try:
                self.load_args_from_json_string(self.command_line)
            except:
                raise Exception("JSON not supplied, did you use the popup?")


    async def parse_dictionary(self, dictionary):
        self.load_args_from_dictionary(dictionary)


class PortScanCommand(CommandBase):
    cmd = "portscan"
    needs_admin = False
    help_cmd = "portscan"
    description = "Scan host(s) for open ports."
    version = 1
    author = "@djhohnstein"
    argument_class = PortScanArguments
    attackmapping = ["T1046"]
    browser_script = BrowserScript(script_name="portscan_new", author="@djhohnstein", for_new_ui=True)

    async def create_tasking(self, task: MythicTask) -> MythicTask:
        return task

    async def process_response(self, response: AgentResponse):
        pass
