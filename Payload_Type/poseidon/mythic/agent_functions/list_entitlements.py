from mythic_payloadtype_container.MythicCommandBase import *
import json


class ListEntitlementsArguments(TaskArguments):
    def __init__(self, command_line, **kwargs):
        super().__init__(command_line, **kwargs)
        self.args = [
            CommandParameter(
                name="pid",
                display_name="Pid to query (-1 for all)",
                type=ParameterType.Number,
                default_value=-1,
                description="PID of process to query (-1 for all)",
                parameter_group_info=[ParameterGroupInfo(
                    required=False
                )]
            )
        ]

    async def parse_arguments(self):
        self.load_args_from_json_string(self.command_line)

    async def parse_dictionary(self, dictionary):
        self.load_args_from_dictionary(dictionary)


class ListEntitlementCommand(CommandBase):
    cmd = "list_entitlements"
    needs_admin = False
    help_cmd = "list_entitlements"
    description = "Use CSOps Syscall to list the entitlements for processes (-1 for all processes)"
    version = 1
    author = "@its_a_feature_"
    argument_class = ListEntitlementsArguments
    attackmapping = ["T1057"]
    browser_script = BrowserScript(script_name="list_entitlements_new", author="@its_a_feature_", for_new_ui=True)
    attributes = CommandAttributes(
        # uncomment when poseidon can dynamically compile commands
        supported_os=[SupportedOS.MacOS]
    )

    async def create_tasking(self, task: MythicTask) -> MythicTask:
        if task.args.get_arg("pid") == -1:
            task.display_params = " for all running processes"
        else:
            task.display_params = " for pid " + str(task.args.get_arg("pid"))
        return task

    async def process_response(self, response: AgentResponse):
        pass
